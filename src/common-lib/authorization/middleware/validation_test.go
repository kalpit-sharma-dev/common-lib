package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/auth"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/middleware"
	mock2 "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/middleware/mock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token/mock"
)

func TestTokenValidation_Handler(t *testing.T) {
	var (
		ctrl             = gomock.NewController(t)
		exchangerMock    = mock.NewMockExchanger(ctrl)
		sigValidatorMock = mock.NewMockValidator(ctrl)
		cacheMock        = mock2.NewMockCache(ctrl)

		cfg      = token.AuthorizationConfig{URL: "http://127.0.0.1"}
		pid      = "pid"
		uid      = "uid"
		jwtToken = "some.jwt.token"

		target = middleware.NewTokenValidation(cfg, exchangerMock, sigValidatorMock, nil, cacheMock, nil)
	)

	t.Run("positive", func(t *testing.T) {
		request := &http.Request{Header: http.Header{auth.AuthorizationHeader: []string{jwtToken}}}

		ctx := context.WithValue(context.Background(), auth.PartnerIDKey, pid)
		ctx = context.WithValue(ctx, auth.UserIDKey, uid)

		request = request.WithContext(ctx)
		recorder := httptest.NewRecorder()

		sigValidatorMock.EXPECT().Validate(ctx, jwtToken, pid, uid).Return(nil)

		next := func(rw http.ResponseWriter, r *http.Request) {
			assert.Equal(t, uid, r.Context().Value(auth.UserIDKey).(string))
		}

		target.Handler(next)(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatal("unexpected status code", recorder.Code)
		}
	})

	t.Run("negative_validation_failed", func(t *testing.T) {
		request := &http.Request{Header: http.Header{auth.AuthorizationHeader: []string{jwtToken}}}

		ctx := context.WithValue(context.Background(), auth.PartnerIDKey, pid)
		ctx = context.WithValue(ctx, auth.UserIDKey, uid)

		request = request.WithContext(ctx)
		recorder := httptest.NewRecorder()

		sigValidatorMock.EXPECT().Validate(ctx, jwtToken, pid, uid).Return(errors.New(""))

		target.Handler(nil)(recorder, request)
		if recorder.Code != http.StatusForbidden {
			t.Fatal("unexpected status code", recorder.Code)
		}
	})

	t.Run("negative_no_token", func(t *testing.T) {
		request := &http.Request{Header: http.Header{}}
		recorder := httptest.NewRecorder()

		target.Handler(nil)(recorder, request)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatal("unexpected status code", recorder.Code)
		}
	})

	t.Run("negative_no_partner", func(t *testing.T) {
		request := &http.Request{Header: http.Header{}}
		request.AddCookie(&http.Cookie{Name: auth.JWTCookie, Value: jwtToken})

		ctx := context.WithValue(context.Background(), auth.SessionTokenKey, "token")
		request = request.WithContext(ctx)
		recorder := httptest.NewRecorder()

		target.Handler(nil)(recorder, request)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatal("unexpected status code", recorder.Code)
		}
	})

	t.Run("negative_no_user", func(t *testing.T) {
		request := &http.Request{Header: http.Header{}}
		request.AddCookie(&http.Cookie{Name: auth.JWTCookie, Value: jwtToken})

		ctx := context.WithValue(context.Background(), auth.PartnerIDKey, pid)
		request = request.WithContext(ctx)

		recorder := httptest.NewRecorder()

		target.Handler(nil)(recorder, request)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatal("unexpected status code", recorder.Code)
		}
	})

	t.Run("negative_no_partner", func(t *testing.T) {
		request := &http.Request{Header: http.Header{}}
		request.AddCookie(&http.Cookie{Name: auth.JWTCookie, Value: jwtToken})

		ctx := context.WithValue(context.Background(), auth.SessionTokenKey, "token")
		request = request.WithContext(ctx)

		recorder := httptest.NewRecorder()

		target.Handler(nil)(recorder, request)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatal("unexpected status code", recorder.Code)
		}
	})

	t.Run("positive_exchange", func(t *testing.T) {
		request := &http.Request{Header: http.Header{}}

		ctx := context.WithValue(context.Background(), auth.PartnerIDKey, pid)
		ctx = context.WithValue(ctx, auth.UserIDKey, uid)
		ctx = context.WithValue(ctx, auth.SessionTokenKey, "token")

		request = request.WithContext(ctx)
		recorder := httptest.NewRecorder()

		exchangerMock.EXPECT().Exchange(ctx, "token", pid).Return("jwt", nil)
		cacheMock.EXPECT().Get(gomock.Any()).Return(nil, errors.New(""))
		cacheMock.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any())

		next := func(rw http.ResponseWriter, r *http.Request) {
			assert.Equal(t, uid, r.Context().Value(auth.UserIDKey).(string))
		}

		target.Handler(next)(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatal("unexpected status code", recorder.Code)
		}
	})

	t.Run("negative_exchange", func(t *testing.T) {
		request := &http.Request{Header: http.Header{}}

		ctx := context.WithValue(context.Background(), auth.PartnerIDKey, pid)
		ctx = context.WithValue(ctx, auth.UserIDKey, uid)
		ctx = context.WithValue(ctx, auth.SessionTokenKey, "token")

		request = request.WithContext(ctx)
		recorder := httptest.NewRecorder()

		exchangerMock.EXPECT().Exchange(ctx, "token", pid).Return("", token.ErrInvalidAuthzResponse)
		cacheMock.EXPECT().Get(gomock.Any()).Return(nil, errors.New(""))

		target.Handler(nil)(recorder, request)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatal("unexpected status code", recorder.Code)
		}
	})
}

func TestTrustingTokenValidation_Handler(t *testing.T) {
	var (
		ctrl             = gomock.NewController(t)
		sigValidatorMock = mock.NewMockValidator(ctrl)
		cacheMock        = mock2.NewMockCache(ctrl)

		target = middleware.NewTrustingTokenValidation(token.AuthorizationConfig{URL: "http://127.0.0.1"}, sigValidatorMock, nil, cacheMock, nil)
		jwt    = "some.jwt.token"
	)

	t.Run("requests_with_jwts_work", func(t *testing.T) {
		request := &http.Request{Header: http.Header{auth.AuthorizationHeader: []string{jwt}}}
		recorder := httptest.NewRecorder()

		sigValidatorMock.EXPECT().Validate(request.Context(), jwt, "", "").Return(nil)

		next := func(rw http.ResponseWriter, r *http.Request) {
			assert.False(t, r.Context().Value(auth.InternalRequest).(bool))
		}

		target.Middleware(http.HandlerFunc(next)).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatal("unexpected status code", recorder.Code)
		}
	})

	t.Run("bad_jwts_fail", func(t *testing.T) {
		request := &http.Request{Header: http.Header{auth.AuthorizationHeader: []string{jwt}}}
		recorder := httptest.NewRecorder()

		sigValidatorMock.EXPECT().Validate(request.Context(), jwt, "", "").Return(errors.New(""))

		target.Middleware(http.HandlerFunc(nil)).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusForbidden {
			t.Fatal("unexpected status code", recorder.Code)
		}
	})

	t.Run("internal_requests_work", func(t *testing.T) {
		request := &http.Request{Header: http.Header{}}
		recorder := httptest.NewRecorder()

		next := func(rw http.ResponseWriter, r *http.Request) {
			assert.True(t, r.Context().Value(auth.InternalRequest).(bool))
		}

		target.Middleware(http.HandlerFunc(next)).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatal("unexpected status code", recorder.Code)
		}
	})
}
