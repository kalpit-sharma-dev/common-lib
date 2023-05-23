package ratelimit

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
)

const (
	headerLimit       = "X-RateLimit-Limit"
	headerRemaining   = "X-RateLimit-Remaining"
	headerContentType = "Content-Type"
	applicationJSON   = "application/json"
)

type (
	enabledStatusView struct {
		Enabled bool `json:"enabled"`
	}

	errorResponseView struct {
		Error string `json:"error"`
	}
)

// SetConfig updates the rate limit configuration
func SetConfig(w http.ResponseWriter, r *http.Request) {
	if limiterMW == nil {
		sendInternalServerError(w, errNilLimiter, transactionID(r))
		return
	}
	limiterMW.setConfig(w, r)
}

// GetConfig returns the rate limit configuration
func GetConfig(w http.ResponseWriter, r *http.Request) {
	if limiterMW == nil {
		sendInternalServerError(w, errNilLimiter, transactionID(r))
		return
	}
	limiterMW.getConfig(w, r)
}

// SetEnabled enable or disable the limiter
func SetEnabled(w http.ResponseWriter, r *http.Request) {
	if limiterMW == nil {
		sendInternalServerError(w, errNilLimiter, transactionID(r))
		return
	}
	limiterMW.setEnabled(w, r)
}

func transactionID(r *http.Request) string {
	return utils.GetTransactionIDFromRequest(r)
}

func sendInternalServerError(w http.ResponseWriter, err error, transactionID string) {
	logError(transactionID, http.StatusText(http.StatusInternalServerError), err)
	w.WriteHeader(http.StatusInternalServerError)
}

func sendBadRequest(w http.ResponseWriter, err error, transactionID string) {
	logError(transactionID, http.StatusText(http.StatusBadRequest), err)
	renderJSON(w, http.StatusBadRequest, errorResponseView{Error: err.Error()}, transactionID)
}

func sendOK(w http.ResponseWriter, response interface{}, transactionID string) {
	renderJSON(w, http.StatusOK, response, transactionID)
}

func renderJSON(w http.ResponseWriter, status int, response interface{}, transactionID string) {
	data, err := json.Marshal(response)
	if err != nil {
		sendInternalServerError(w, err, transactionID)
		return
	}
	w.Header().Set(headerContentType, applicationJSON)
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func limitsHeader(w http.ResponseWriter, limit, remaining int64) {
	w.Header().Set(headerLimit, fmt.Sprintf("%d", limit))
	w.Header().Set(headerRemaining, fmt.Sprintf("%d", remaining))
}
