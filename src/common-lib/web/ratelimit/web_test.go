package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockMiddlewareNoContent struct{}

func TestGetConfig(t *testing.T) {
	defer mockLimiterMW(&mockMiddlewareNoContent{})()
	r := httptest.NewRequest(http.MethodGet, mockTestURL, nil)
	w := httptest.NewRecorder()

	GetConfig(w, r)

	require.NotNil(t, limiterMW)
	require.Equal(t, http.StatusNoContent, w.Result().StatusCode)
}

func TestGetConfig_WithInternalServerError(t *testing.T) {
	defer mockLimiterMW(nil)()
	r := httptest.NewRequest(http.MethodGet, mockTestURL, nil)
	w := httptest.NewRecorder()

	GetConfig(w, r)

	require.Nil(t, limiterMW)
	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestSetConfig(t *testing.T) {
	defer mockLimiterMW(&mockMiddlewareNoContent{})()
	r := httptest.NewRequest(http.MethodPut, mockTestURL, nil)
	w := httptest.NewRecorder()

	SetConfig(w, r)

	require.NotNil(t, limiterMW)
	require.Equal(t, http.StatusNoContent, w.Result().StatusCode)
}

func TestSetConfig_WithInternalServerError(t *testing.T) {
	defer mockLimiterMW(nil)()
	r := httptest.NewRequest(http.MethodPut, mockTestURL, nil)
	w := httptest.NewRecorder()

	SetConfig(w, r)

	require.Nil(t, limiterMW)
	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestSetEnabled(t *testing.T) {
	defer mockLimiterMW(&mockMiddlewareNoContent{})()
	r := httptest.NewRequest(http.MethodPatch, mockTestURL, nil)
	w := httptest.NewRecorder()

	SetEnabled(w, r)

	require.NotNil(t, limiterMW)
	require.Equal(t, http.StatusNoContent, w.Result().StatusCode)
}

func TestSetEnabled_WithInternalServerError(t *testing.T) {
	defer mockLimiterMW(nil)()
	r := httptest.NewRequest(http.MethodPatch, mockTestURL, nil)
	w := httptest.NewRecorder()

	SetEnabled(w, r)

	require.Nil(t, limiterMW)
	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func mockLimiterMW(mw middleware) func() {
	orig := limiterMW
	limiterMW = mw
	return func() {
		limiterMW = orig
	}
}

func (m mockMiddlewareNoContent) CheckRateLimit(_ http.HandlerFunc, _, _ string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (m mockMiddlewareNoContent) setConfig(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (m mockMiddlewareNoContent) getConfig(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (m mockMiddlewareNoContent) setEnabled(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
