// Deprecated: this package isn't deprecated, but is unstable and does NOT follow semver. If you use this package, upgrade platform-common-lib with caution.
package entityreference

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gocql/gocql"
	"gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/auth"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

const (
	// Hard reference type
	Hard Type = "hard"
	// Soft reference type
	Soft Type = "soft"

	defaultUID = "cleanup"
)

// Type is a type of the reference
type Type string

// Reference is a db model
type Reference struct {
	PartnerID               string     `db:"partner_id" cql:"partner_id"`
	EntityID                gocql.UUID `db:"entity_id" cql:"entity_id"`
	Service                 string     `db:"service" cql:"service"`
	ReferencingObjectID     gocql.UUID `db:"referencing_object_id" cql:"referencing_object_id"`
	Type                    Type       `db:"type" cql:"type"`
	ValidationCallbackURL   string     `db:"validation_callback_url" cql:"validation_callback_url"`
	NotificationCallbackURL string     `db:"notification_callback_url" cql:"notification_callback_url"`
}

// ReferenceRequest is a request model
type ReferenceRequest struct {
	ReferencingObjectID     gocql.UUID `json:"referencing_object_id" validate:"required"`
	Type                    Type       `json:"type" validate:"required,oneof=hard soft"`
	ValidationCallbackURL   string     `json:"validation_callback_url" validate:"required"`
	NotificationCallbackURL string     `json:"notification_callback_url" validate:"required_if=Type soft"`
}

// ReferenceResponse is a response model
type ReferenceResponse struct {
	EntityID            gocql.UUID `json:"entity_id"`
	ReferencingObjectID gocql.UUID `json:"referencing_object_id"`
	Service             string     `json:"service"`
	Type                Type       `json:"type"`
}

// NotificationMessage is a notification model
type NotificationMessage struct {
	EntityID            gocql.UUID `json:"entity_id"`
	ReferencingObjectID gocql.UUID `json:"referencing_object_id"`
}

// Repo is a repository interface
type Repo interface {
	AddOne(item *Reference) error
	GetReferences(keys ...interface{}) ([]*Reference, error)
	GetOne(keyCols ...interface{}) (*Reference, error)
	DeleteOne(item *Reference) error
}

// Client is an interface for http.Client
type Client interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// ManagementUsecase is a reference manager usecase struct
type ManagementUsecase struct {
	repo   Repo
	client Client
	log    logger.Log
}

// NewManagementUsecase is a constructor
func NewManagementUsecase(repo Repo, client Client, log logger.Log) *ManagementUsecase {
	return &ManagementUsecase{repo: repo, client: client, log: log}
}

// Create reference
func (m *ManagementUsecase) Create(
	_ context.Context,
	reference *ReferenceRequest,
	entityID gocql.UUID,
	serviceName, partnerID string,
) error {
	referenceDB := &Reference{
		PartnerID:               partnerID,
		EntityID:                entityID,
		Service:                 serviceName,
		ReferencingObjectID:     reference.ReferencingObjectID,
		Type:                    reference.Type,
		ValidationCallbackURL:   reference.ValidationCallbackURL,
		NotificationCallbackURL: reference.NotificationCallbackURL,
	}
	return m.repo.AddOne(referenceDB)
}

// Get one reference
func (m *ManagementUsecase) Get(
	_ context.Context,
	entityID, referenceID gocql.UUID,
	serviceName, partnerID string,
) (*ReferenceResponse, error) {
	reference, err := m.repo.GetOne(partnerID, entityID, serviceName, referenceID)
	if err != nil {
		return nil, err
	}

	response := &ReferenceResponse{
		EntityID:            reference.EntityID,
		ReferencingObjectID: reference.ReferencingObjectID,
		Service:             reference.Service,
		Type:                reference.Type,
	}
	return response, nil
}

// GetAll references
func (m *ManagementUsecase) GetAll(_ context.Context, entityID gocql.UUID, serviceName, partnerID string) ([]*ReferenceResponse, error) {
	var (
		references []*Reference
		err        error
	)

	if serviceName != "" {
		references, err = m.repo.GetReferences(partnerID, entityID, serviceName)
	} else {
		references, err = m.repo.GetReferences(partnerID, entityID)
	}
	if err != nil {
		return nil, err
	}

	response := make([]*ReferenceResponse, 0, len(references))
	for _, reference := range references {
		response = append(response, &ReferenceResponse{
			EntityID:            reference.EntityID,
			ReferencingObjectID: reference.ReferencingObjectID,
			Service:             reference.Service,
			Type:                reference.Type,
		})
	}
	return response, nil
}

// Delete reference
func (m *ManagementUsecase) Delete(_ context.Context, entityID, referenceID gocql.UUID, serviceName, partnerID string) error {
	reference := Reference{
		PartnerID:           partnerID,
		EntityID:            entityID,
		Service:             serviceName,
		ReferencingObjectID: referenceID,
	}
	return m.repo.DeleteOne(&reference)
}

// CleanUp all references which are not valid
func (m *ManagementUsecase) CleanUp(ctx context.Context) error {
	references, err := m.repo.GetReferences()
	if err != nil {
		m.log.Error(getTransactionID(ctx), "failed to get references", err.Error())
		return err
	}

	return m.cleanup(ctx, references)
}

// CleanUpByPartner cleanup partner specific
func (m *ManagementUsecase) CleanUpByPartner(ctx context.Context, partnerID string) error {
	references, err := m.repo.GetReferences(partnerID)
	if err != nil {
		m.log.Error(getTransactionID(ctx), "failed to get references", err.Error())
		return err
	}

	return m.cleanup(ctx, references)
}

func (m *ManagementUsecase) cleanup(ctx context.Context, references []*Reference) error {
	var err error
	for _, reference := range references {
		err = m.makeRequest(ctx, http.MethodGet, reference.ValidationCallbackURL, nil)
		if err == nil {
			continue
		}
		if errors.As(err, &NotFoundError{}) {
			if err = m.repo.DeleteOne(reference); err != nil {
				m.log.Error(
					getTransactionID(ctx),
					"failed to delete reference",
					"partnerID: %s, entityID: %s, referencingObjectID: %s, err: %s",
					reference.PartnerID, reference.EntityID, reference.ReferencingObjectID.String(), err)
			}
		} else {
			m.log.Warn(
				getTransactionID(ctx),
				"failed to validate reference via callbackURL",
				"partnerID: %s, entityID: %s, referencingObjectID: %s, callbackURL: %s, err: %s",
				reference.PartnerID, reference.EntityID, reference.ReferencingObjectID.String(), reference.ValidationCallbackURL, err,
			)
		}
	}

	return nil
}

// ValidateDeletion validate possibility to delete entity
func (m *ManagementUsecase) ValidateDeletion(ctx context.Context, references []*Reference) error {
	var err error
	for _, reference := range references {
		if reference.Type == Hard {
			return NewConflictReferenceError(references)
		}
	}

	for _, reference := range references {
		if reference.Type == Soft {
			if err = m.NotifySoftReference(ctx, reference); err != nil {
				m.log.Error(
					getTransactionID(ctx),
					"failed to send notification",
					"partnerID: %s, entityID: %s, referenceID: %s, notifyURL: %s, err: %s",
					reference.PartnerID, reference.EntityID.String(),
					reference.ReferencingObjectID.String(), reference.NotificationCallbackURL, err)
			}
		}
		if err = m.repo.DeleteOne(reference); err != nil {
			return err
		}
	}
	return nil
}

// NotifySoftReference when reference was deleted
func (m *ManagementUsecase) NotifySoftReference(ctx context.Context, reference *Reference) error {
	message := NotificationMessage{
		EntityID:            reference.EntityID,
		ReferencingObjectID: reference.ReferencingObjectID,
	}

	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return m.makeRequest(ctx, http.MethodPost, reference.NotificationCallbackURL, msg)
}

func (m *ManagementUsecase) makeRequest(ctx context.Context, method, url string, body []byte) error {
	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	request.Header.Add(auth.TransactionHeader, getTransactionID(ctx))
	request.Header.Add(auth.UserIDHeader, defaultUID)

	response, err := m.client.Do(ctx, request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return parseErrResponse(response.Body, response.StatusCode)
	}
	return nil
}

func parseErrResponse(reader io.Reader, statusCode int) error {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(reader); err != nil {
		return err
	}

	switch statusCode {
	case http.StatusNotFound:
		return NotFoundError{}
	default:
		return MsgError{
			Message: fmt.Sprintf("request failed with status: %d", statusCode),
			Desc:    buf.String(),
		}
	}
}

// getTransactionID from context value
func getTransactionID(ctx context.Context) string {
	v := ctx.Value(auth.TransactionKey)
	if v == nil {
		return ""
	}

	return v.(string)
}
