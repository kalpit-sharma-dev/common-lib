// Package contextutil provides a standard means of storing and accessing common
// transactional data in our platform within the scope of a single request.
package contextutil

import (
	"context"

	"github.com/google/uuid"
)

type correlationIDType int

// ContextDataKey is the key the data structure is stored as in the context.
const ContextDataKey correlationIDType = iota

// Key values that can be used for each value stored in the context data.
const (
	PartnerID     string = "PartnerID"
	VendorID      string = "VendorID"
	ProductID     string = "ProductID"
	UserID        string = "UserID"
	CompanyID     string = "CompanyID"
	SiteID        string = "SiteID"
	AgentID       string = "AgentID"
	EndpointID    string = "EndpointID"
	ClientID      string = "ClientID"
	RequestID     string = "RequestID"
	TransactionID string = "TransactionID"
)

type contextData struct {
	// The unique identifier of a tenant in our system, our partners.
	PartnerID string
	// The unique identifier of a vendor in our system.
	VendorID string
	// The unique identifier of a productID in our system.
	ProductID string
	// The unique identifier of the user that made the request.
	UserID string
	// The unique identifier of a partner's company.
	CompanyID string
	// The unique identifier of a company's site.
	SiteID string
	// The unique identifier of a site's endpoint.
	EndpointID string
	// The unique identifier for a particular agent serving an endpoint.
	AgentID string
	// The client, such as the calling software, making this request.
	// Not to be confused with the Command legacy reference to a "client" which is instead represented by "CompanyId".
	ClientID string
	// A unique identifier for a distributed transaction, passed to each system involved in the request. Also known as a correlation ID.
	TransactionID string
	// A unique identifier scoped within a single part of a request contained to only this system.
	// Not to be confused with transaction ID which can represent a request across multiple systems.
	// If "System A" is called twice within the scope of a single transaction,
	// each request to System A would have a unique request ID, but they would share a transaction ID across both requests.
	// This will default to a new GUID when NewContextData is called.
	RequestID string
}

// NewContextData construct new contextData with transactionID, partnerID, and userID.
func NewContextData(transactionID, partnerID, userID string) (contextData, error) {
	c := contextData{
		PartnerID:     partnerID,
		TransactionID: transactionID,
		UserID:        userID,
		RequestID:     uuid.New().String(),
	}

	return c, nil
}

// WithValue returns a context which knows all context data.
func WithValue(ctx context.Context, contextDataS contextData) context.Context {
	return context.WithValue(ctx, ContextDataKey, contextDataS)
}

// GetData get context data with contextDataKey.
func GetData(ctx context.Context) contextData {
	var ctxData contextData
	if ctx != nil {
		ctxData, _ = ctx.Value(ContextDataKey).(contextData)
	}

	return ctxData
}
