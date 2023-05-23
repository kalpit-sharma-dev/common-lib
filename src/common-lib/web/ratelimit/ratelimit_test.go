package ratelimit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/ratelimit/mock"
)

const mockTestURL = "/test-url"

type mockResponder struct{}

func TestInit(t *testing.T) {
	defer mockLimiterMW(nil)()
	defer mockInMemoryConfig(nil)()
	cfg := mockConfig()
	store, finish := mockStorageWithConfig(t, cfg, 1, 0)
	defer finish()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	limiter := &middlewareImpl{}

	require.Nil(t, limiterMW)

	err := Init(ctx, mockResponder{}, store, cfg)
	got := Middleware()

	require.NoError(t, err)
	require.IsType(t, limiter, got)
}

func TestInit_WithError(t *testing.T) {
	cfg := mockConfig()
	store, finish := mockStorage(t)
	defer finish()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Responder is nil", func(t *testing.T) {
		err := Init(ctx, nil, store, cfg)
		t.Log(err)
		require.Error(t, err)
	})
	t.Run("Storage is nil", func(t *testing.T) {
		err := Init(ctx, mockResponder{}, nil, cfg)
		t.Log(err)
		require.Error(t, err)
	})
}

func Test_middlewareImpl_CheckRateLimit(t *testing.T) {
	cfg := mockConfig()
	cfgLimit := cfg.Groups[mockGroup].Overrides[mockKey]
	mockCount := int64(1)
	expLimitHeader := fmt.Sprintf("%d", cfgLimit)
	expRemainingHeader := fmt.Sprintf("%d", cfgLimit-mockCount)

	store, finish := mockStorageWithConfig(t, cfg, 0, 1)
	defer finish()

	handler := setupLimiter(store, mockCountGetter(mockCount, nil))
	r := httptest.NewRequest(http.MethodGet, mockTestURL, nil)
	w := httptest.NewRecorder()

	handler(w, r)

	res := w.Result()

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, expLimitHeader, res.Header.Get(headerLimit))
	require.Equal(t, expRemainingHeader, res.Header.Get(headerRemaining))
}

func Test_middlewareImpl_CheckRateLimit_WithBadRequest(t *testing.T) {
	store, finish := mockStorageWithConfig(t, mockConfig(), 0, 1)
	defer finish()
	handler := setupLimiter(store, mockCountGetter(0, errors.New("something wrong")))
	r := httptest.NewRequest(http.MethodGet, mockTestURL, nil)
	w := httptest.NewRecorder()

	handler(w, r)

	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func Test_middlewareImpl_CheckRateLimit_WithTooManyRequests(t *testing.T) {
	cfg := mockConfig()
	cfg.Groups[mockGroup].Overrides[mockKey] = 1
	store, finish := mockStorageWithConfig(t, cfg, 0, 1)
	defer finish()
	handler := setupLimiter(store, mockCountGetter(10, nil))
	r := httptest.NewRequest(http.MethodGet, mockTestURL, nil)
	w := httptest.NewRecorder()

	handler(w, r)

	require.Equal(t, http.StatusTooManyRequests, w.Result().StatusCode)
}

func Test_middlewareImpl_getConfig(t *testing.T) {
	cfg := mockConfig()
	store, finish := mockStorageWithConfig(t, cfg, 0, 1)
	defer finish()
	limiter := &middlewareImpl{storage: store}
	r := httptest.NewRequest(http.MethodGet, mockTestURL, nil)
	w := httptest.NewRecorder()

	limiter.getConfig(w, r)

	res := w.Result()
	var got Config
	_ = json.NewDecoder(res.Body).Decode(&got)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, cfg, got)
}

func Test_middlewareImpl_getConfig_WithInternalServerError(t *testing.T) {
	store, finish := mockStorage(t)
	defer finish()
	defer mockInMemoryConfig(nil)()
	store.EXPECT().Get(configStorageKey).Return(nil, errors.New("storage error"))
	limiter := &middlewareImpl{storage: store}
	r := httptest.NewRequest(http.MethodGet, mockTestURL, nil)
	w := httptest.NewRecorder()

	limiter.getConfig(w, r)

	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func Test_middlewareImpl_setConfig(t *testing.T) {
	cfg := mockConfig()
	store, finish := mockStorageWithConfig(t, cfg, 1, 0)
	defer finish()
	limiter := &middlewareImpl{storage: store}
	b, _ := json.Marshal(cfg)
	r := httptest.NewRequest(http.MethodPut, mockTestURL, bytes.NewReader(b))
	w := httptest.NewRecorder()

	limiter.setConfig(w, r)

	res := w.Result()
	var got Config
	_ = json.NewDecoder(res.Body).Decode(&got)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, cfg, got)
}

func Test_middlewareImpl_setConfig_WithBadRequest(t *testing.T) {
	cfg := mockConfig()
	cfg.Algorithm = "invalidAlgorithm"
	store, finish := mockStorage(t)
	defer finish()
	limiter := &middlewareImpl{storage: store}
	b, _ := json.Marshal(cfg)
	r := httptest.NewRequest(http.MethodPut, mockTestURL, bytes.NewReader(b))
	w := httptest.NewRecorder()

	limiter.setConfig(w, r)

	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func Test_middlewareImpl_setEnabled(t *testing.T) {
	defer mockInMemoryConfig(nil)()

	cfg := mockConfig()
	cfg.Enabled = false

	getCfg, _ := json.Marshal(cfg)
	setCfg, _ := json.Marshal(mockConfig())

	store, finish := mockStorage(t)
	store.EXPECT().Get(configStorageKey).Return(string(getCfg), nil)
	store.EXPECT().Set(configStorageKey, setCfg).Return(nil)
	defer finish()

	limiter := &middlewareImpl{storage: store}

	expected := enabledStatusView{Enabled: true}
	b, _ := json.Marshal(expected)
	r := httptest.NewRequest(http.MethodPatch, mockTestURL, bytes.NewReader(b))
	w := httptest.NewRecorder()

	limiter.setEnabled(w, r)

	res := w.Result()
	var got enabledStatusView
	_ = json.NewDecoder(res.Body).Decode(&got)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, expected, got)
}

func Test_middlewareImpl_setEnabled_WithBadRequest(t *testing.T) {
	store, finish := mockStorage(t)
	defer finish()

	limiter := &middlewareImpl{storage: store}

	r := httptest.NewRequest(http.MethodPatch, mockTestURL, nil)
	w := httptest.NewRecorder()

	limiter.setEnabled(w, r)

	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func setupLimiter(store Storage, countFn countGetter) http.HandlerFunc {
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	limiter := &middlewareImpl{
		storage: store,
		rp:      mockResponder{},
		count:   countFn,
	}
	return limiter.CheckRateLimit(next, mockGroup, mockKey)
}

func mockStorageWithConfig(t *testing.T, cfg Config, setTimes, getTimes int) (*mock.MockStorage, func()) {
	ctrl := gomock.NewController(t)
	b, _ := json.Marshal(cfg)
	m := mock.NewMockStorage(ctrl)
	m.EXPECT().Set(configStorageKey, b).Return(nil).Times(setTimes)
	m.EXPECT().Get(configStorageKey).Return(string(b), nil).Times(getTimes)

	restore := mockInMemoryConfig(nil)

	return m, func() {
		ctrl.Finish()
		restore()
	}
}

func mockCountGetter(count int64, err error) countGetter {
	return func(_ Storage, _ countParams) (int64, error) {
		return count, err
	}
}

func (mr mockResponder) RespondRateLimit(rp Response, _ *http.Request, w http.ResponseWriter) {
	w.WriteHeader(rp.StatusCode)
}
