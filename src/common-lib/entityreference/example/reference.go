package example

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/auth"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/entityreference"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

type YourRepo struct{}

func (r *YourRepo) AddOne(item *entityreference.Reference) error {
	// implement that wrapper method in your service
	return nil
}

func (r *YourRepo) GetReferences(keys ...interface{}) ([]*entityreference.Reference, error) {
	// implement that wrapper method in your service
	return []*entityreference.Reference{}, nil
}

func (r *YourRepo) GetOne(keyCols ...interface{}) (*entityreference.Reference, error) {
	// implement that wrapper method in your service
	return &entityreference.Reference{}, nil
}

func (r *YourRepo) DeleteOne(item *entityreference.Reference) error {
	// implement that wrapper method in your service
	return nil
}

func ExampleGetReferenceHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	service := vars["service"]
	partnerID := vars["partnerID"]

	entityID, err := gocql.ParseUUID(vars["entityID"])
	if err != nil {
		return
	}
	referenceID, err := gocql.ParseUUID(vars["reference"])
	if err != nil {
		return
	}

	yourRepo := new(YourRepo)
	ctx := context.WithValue(r.Context(), auth.TransactionKey, "txID")
	usecase := entityreference.NewManagementUsecase(yourRepo, nil, logger.DiscardLogger())
	response, err := usecase.Get(ctx, entityID, referenceID, service, partnerID)
	if err != nil {
		return
	}

	marshaled, _ := json.Marshal(response)
	w.Write(marshaled)
}

func ExampleCreateReferenceHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	service := vars["service"]
	partnerID := vars["partnerID"]

	entityID, err := gocql.ParseUUID(vars["entityID"])
	if err != nil {
		return
	}

	request := &entityreference.ReferenceRequest{}
	err = json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		return
	}
	defer r.Body.Close()

	ctx := context.WithValue(r.Context(), auth.TransactionKey, "txID")

	err = validator.New().StructCtx(ctx, request)
	if err != nil {
		return
	}

	yourRepo := new(YourRepo)
	usecase := entityreference.NewManagementUsecase(yourRepo, nil, logger.DiscardLogger())
	err = usecase.Create(ctx, request, entityID, service, partnerID)
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
}
