package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/auth"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token/permission"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

func TestPermission_AssertHandler(t *testing.T) {
	var validJWT = `eyJhbGciOiJSUzUxMiIsImtpZCI6ImUwM2RiODEwLTJmMTctNGE0OC1iYTVkLTc2NGU3NzljMjc3NyIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTAxNjU5MjkiLCJwYXJ0bmVyX2lkIjoiNTAwMDE3OTQiLCJwZXJtaXNzaW9ucyI6WyJSTU1TZXR1cC5Fc3NlbnRpYWxzLkVkaXRNZW1iZXJOZXcuRnVsbCIsIkRldmljZS5NYW5hZ2UiXSwiZXhwIjoxNTk4MzcyODkwLCJpYXQiOjE1OTgzNjU2OTAsImlzcyI6IkF1dGhvcml6YXRpb24gTVMifQ.HgZqSiHOKeYrvtcHH4T9HmQlcha0wAMCmIW62N4L3YyQy5bm-CZR9fwJkRRHEruLiPSnkqh8bpO8IQ23QBV6hs29BGKXwSgUx5V1YerQ2Ek_gJdggHqz_VC-S24Ifs6Eus30jDFAOL-O5YPqWzy8LPurZWum_hzhnUnnbeePfc0`

	log, err := logger.Create(logger.Config{Name: "permission_test_1", Destination: logger.DISCARD})
	if err != nil {
		t.Fatal(err)
	}
	var target = NewPermission(nil, log)

	testCases := []struct {
		name           string
		permissions    []string
		next           func(rw http.ResponseWriter, r *http.Request)
		validationType permission.ValidationType
		jwt            string
		buildRequest   func(jwt string) *http.Request
		statusCode     int
	}{
		{
			name: "Invalid_JWT",
			jwt:  "invalid_jwt",
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r
			},
			statusCode: http.StatusUnauthorized,
		},
		{
			name: "No_JWT",
			buildRequest: func(jwt string) *http.Request {
				return &http.Request{Header: http.Header{}}
			},
			statusCode: http.StatusUnauthorized,
		},
		{
			name:           "ValidToken_Forbidden",
			permissions:    []string{"read"},
			validationType: permission.AnyOf,
			jwt:            validJWT,
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r
			},
			statusCode: http.StatusForbidden,
		},
		{
			name:        "ValidToken_Success",
			permissions: []string{"Device.Manage"},
			next: func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
			validationType: permission.AnyOf,
			jwt:            validJWT,
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tCase := range testCases {
		tc := tCase
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			target.AssertHandler(tc.next, tc.permissions, tc.validationType)(recorder, tc.buildRequest(tc.jwt))
			assert.Equal(t, recorder.Code, tc.statusCode)
		})
	}
}
func TestTrustingPermission_AssertHandler(t *testing.T) {
	var validJWT = `eyJhbGciOiJSUzUxMiIsImtpZCI6ImUwM2RiODEwLTJmMTctNGE0OC1iYTVkLTc2NGU3NzljMjc3NyIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTAxNjU5MjkiLCJwYXJ0bmVyX2lkIjoiNTAwMDE3OTQiLCJwZXJtaXNzaW9ucyI6WyJSTU1TZXR1cC5Fc3NlbnRpYWxzLkVkaXRNZW1iZXJOZXcuRnVsbCIsIkRldmljZS5NYW5hZ2UiXSwiZXhwIjoxNTk4MzcyODkwLCJpYXQiOjE1OTgzNjU2OTAsImlzcyI6IkF1dGhvcml6YXRpb24gTVMifQ.HgZqSiHOKeYrvtcHH4T9HmQlcha0wAMCmIW62N4L3YyQy5bm-CZR9fwJkRRHEruLiPSnkqh8bpO8IQ23QBV6hs29BGKXwSgUx5V1YerQ2Ek_gJdggHqz_VC-S24Ifs6Eus30jDFAOL-O5YPqWzy8LPurZWum_hzhnUnnbeePfc0`

	log, err := logger.Create(logger.Config{Name: "trusting_permission_test_1", Destination: logger.DISCARD})
	if err != nil {
		t.Fatal(err)
	}
	var target = NewTrustingPermission(nil, log)

	testCases := []struct {
		name           string
		permissions    []string
		next           func(rw http.ResponseWriter, r *http.Request)
		validationType permission.ValidationType
		jwt            string
		buildRequest   func(jwt string) *http.Request
		statusCode     int
	}{
		{
			name: "Invalid_JWT",
			jwt:  "invalid_jwt",
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r
			},
			statusCode: http.StatusUnauthorized,
		},
		{
			name: "No_JWT",
			buildRequest: func(jwt string) *http.Request {
				return &http.Request{Header: http.Header{}}
			},
			next: func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
			statusCode: http.StatusOK,
		},
		{
			name:           "ValidToken_Forbidden",
			permissions:    []string{"read"},
			validationType: permission.AnyOf,
			jwt:            validJWT,
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r
			},
			statusCode: http.StatusForbidden,
		},
		{
			name:        "ValidToken_Success",
			permissions: []string{"Device.Manage"},
			next: func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
			validationType: permission.AnyOf,
			jwt:            validJWT,
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r
			},
			statusCode: http.StatusOK,
		},
		{
			name:        "ValidToken_AlreadyParsed_Success",
			permissions: []string{"Device.Manage"},
			next: func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
			validationType: permission.AnyOf,
			jwt:            validJWT,
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r.WithContext(context.WithValue(r.Context(), auth.PermissionKey, []string{"Device.Manage"}))
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tCase := range testCases {
		tc := tCase
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			target.CommonAssertHandler(tc.next, tc.validationType, tc.permissions...)(recorder, tc.buildRequest(tc.jwt))
			assert.Equal(t, recorder.Code, tc.statusCode)
		})
	}
}

func TestPermission_DecodeHandler(t *testing.T) {
	var validJWT = `eyJhbGciOiJSUzUxMiIsImtpZCI6ImUwM2RiODEwLTJmMTctNGE0OC1iYTVkLTc2NGU3NzljMjc3NyIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTAxNjU5MjkiLCJwYXJ0bmVyX2lkIjoiNTAwMDE3OTQiLCJwZXJtaXNzaW9ucyI6WyJSTU1TZXR1cC5Fc3NlbnRpYWxzLkVkaXRNZW1iZXJOZXcuRnVsbCIsIkRldmljZS5NYW5hZ2UiXSwiZXhwIjoxNTk4MzcyODkwLCJpYXQiOjE1OTgzNjU2OTAsImlzcyI6IkF1dGhvcml6YXRpb24gTVMifQ.HgZqSiHOKeYrvtcHH4T9HmQlcha0wAMCmIW62N4L3YyQy5bm-CZR9fwJkRRHEruLiPSnkqh8bpO8IQ23QBV6hs29BGKXwSgUx5V1YerQ2Ek_gJdggHqz_VC-S24Ifs6Eus30jDFAOL-O5YPqWzy8LPurZWum_hzhnUnnbeePfc0`
	var validJWTWithoutUserAndPartner = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE2NDM4OTY0NzcsInBlcm1pc3Npb25zIjpbIlJNTVNldHVwLkVzc2VudGlhbHMuRWRpdE1lbWJlck5ldy5GdWxsIiwiRGV2aWNlLk1hbmFnZSJdLCJleHAiOjk2NDM4OTY0NDMsImlzcyI6IkF1dGhvcml6YXRpb24gTVMifQ.kMgbBYFpHHPHVa6HMx1Y7T-BVD6Gq3xHwBlqnof7tCs`

	log, err := logger.Create(logger.Config{Name: "permission_test_2", Destination: logger.DISCARD})
	if err != nil {
		t.Fatal(err)
	}
	var target = NewPermission(nil, log)

	testCases := []struct {
		name         string
		next         func(rw http.ResponseWriter, r *http.Request)
		jwt          string
		buildRequest func(jwt string) *http.Request
		statusCode   int
	}{
		{
			name: "Invalid_JWT",
			jwt:  "invalid_jwt",
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r
			},
			statusCode: http.StatusUnauthorized,
		},
		{
			name: "No_JWT",
			buildRequest: func(jwt string) *http.Request {
				return &http.Request{Header: http.Header{}}
			},
			statusCode: http.StatusUnauthorized,
		},
		{
			name: "ValidToken_Success",
			next: func(rw http.ResponseWriter, r *http.Request) {
				permissions, ok := r.Context().Value(auth.PermissionKey).([]string)
				if !ok {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if len(permissions) == 0 {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if r.Context().Value(auth.UserIDKey) == nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if r.Context().Value(auth.PartnerIDKey) == nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if r.Context().Value(auth.JWTKey) == nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				rw.WriteHeader(http.StatusOK)
			},
			jwt: validJWT,
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r
			},
			statusCode: http.StatusOK,
		},
		{
			name: "ValidToken_without_UserID_and_PartnerID",
			next: func(rw http.ResponseWriter, r *http.Request) {
				permissions, ok := r.Context().Value(auth.PermissionKey).([]string)
				if !ok {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if len(permissions) == 0 {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if r.Context().Value(auth.UserIDKey) != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if r.Context().Value(auth.PartnerIDKey) != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if r.Context().Value(auth.JWTKey) == nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				rw.WriteHeader(http.StatusOK)
			},
			jwt: validJWTWithoutUserAndPartner,
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r
			},
			statusCode: http.StatusOK,
		},
	}
	for _, tCase := range testCases {
		tc := tCase
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			target.DecodeHandler(tc.next)(recorder, tc.buildRequest(tc.jwt))
			require.Equal(t, recorder.Code, tc.statusCode)
		})
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name      string
		headerVal string

		wantJWT string
		wantErr bool
	}{
		{
			name:      "works_with_a_jwt",
			headerVal: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantJWT:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		},
		{
			// This test was added to address code that previously would have failed this test
			name:      "works_with_a_jwt_that_ends_in_a_character_from_the_word_Bearer",
			headerVal: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxOTIzNDU2Nzg5MCIsIm5hbWUiOiJKb2hhbiBEb2UiLCJpYXQiOjE1MTYyMzkyMDIyfQ.HpVro_c5D7c0GDZNj3ep40pul2qmyFW8cIOohTdnp-ge",
			wantJWT:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxOTIzNDU2Nzg5MCIsIm5hbWUiOiJKb2hhbiBEb2UiLCJpYXQiOjE1MTYyMzkyMDIyfQ.HpVro_c5D7c0GDZNj3ep40pul2qmyFW8cIOohTdnp-ge",
		},
		{
			name:      "fails_if_missing_jwt",
			headerVal: "",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			h.Set(auth.AuthorizationHeader, tt.headerVal)
			jwt, err := extractToken(h)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantJWT, jwt)
			}
		})
	}
}

func TestTrustingPermission_DecodeHandler(t *testing.T) {
	var validJWT = `eyJhbGciOiJSUzUxMiIsImtpZCI6ImUwM2RiODEwLTJmMTctNGE0OC1iYTVkLTc2NGU3NzljMjc3NyIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTAxNjU5MjkiLCJwYXJ0bmVyX2lkIjoiNTAwMDE3OTQiLCJwZXJtaXNzaW9ucyI6WyJSTU1TZXR1cC5Fc3NlbnRpYWxzLkVkaXRNZW1iZXJOZXcuRnVsbCIsIkRldmljZS5NYW5hZ2UiXSwiZXhwIjoxNTk4MzcyODkwLCJpYXQiOjE1OTgzNjU2OTAsImlzcyI6IkF1dGhvcml6YXRpb24gTVMifQ.HgZqSiHOKeYrvtcHH4T9HmQlcha0wAMCmIW62N4L3YyQy5bm-CZR9fwJkRRHEruLiPSnkqh8bpO8IQ23QBV6hs29BGKXwSgUx5V1YerQ2Ek_gJdggHqz_VC-S24Ifs6Eus30jDFAOL-O5YPqWzy8LPurZWum_hzhnUnnbeePfc0`
	var validJWTWithoutUserAndPartner = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE2NDM4OTY0NzcsInBlcm1pc3Npb25zIjpbIlJNTVNldHVwLkVzc2VudGlhbHMuRWRpdE1lbWJlck5ldy5GdWxsIiwiRGV2aWNlLk1hbmFnZSJdLCJleHAiOjk2NDM4OTY0NDMsImlzcyI6IkF1dGhvcml6YXRpb24gTVMifQ.kMgbBYFpHHPHVa6HMx1Y7T-BVD6Gq3xHwBlqnof7tCs`

	log, err := logger.Create(logger.Config{Name: "trusting_permission_test_2", Destination: logger.DISCARD})
	if err != nil {
		t.Fatal(err)
	}
	var target = NewTrustingPermission(nil, log)

	testCases := []struct {
		name         string
		next         func(rw http.ResponseWriter, r *http.Request)
		jwt          string
		buildRequest func(jwt string) *http.Request
		statusCode   int
	}{
		{
			name: "Invalid_JWT",
			jwt:  "invalid_jwt",
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r
			},
			statusCode: http.StatusUnauthorized,
		},
		{
			name: "No_JWT",
			buildRequest: func(jwt string) *http.Request {
				return &http.Request{Header: http.Header{}}
			},
			next: func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
			},
			statusCode: http.StatusOK,
		},
		{
			name: "ValidToken_Success",
			next: func(rw http.ResponseWriter, r *http.Request) {
				permissions, ok := r.Context().Value(auth.PermissionKey).([]string)
				if !ok {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if len(permissions) == 0 {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if r.Context().Value(auth.UserIDKey) == nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if r.Context().Value(auth.PartnerIDKey) == nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if r.Context().Value(auth.JWTKey) == nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				rw.WriteHeader(http.StatusOK)
			},
			jwt: validJWT,
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r
			},
			statusCode: http.StatusOK,
		},
		{
			name: "ValidToken_without_UserID_and_PartnerID",
			next: func(rw http.ResponseWriter, r *http.Request) {
				permissions, ok := r.Context().Value(auth.PermissionKey).([]string)
				if !ok {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if len(permissions) == 0 {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if r.Context().Value(auth.UserIDKey) != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if r.Context().Value(auth.PartnerIDKey) != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				if r.Context().Value(auth.JWTKey) == nil {
					rw.WriteHeader(http.StatusInternalServerError)
					return
				}

				rw.WriteHeader(http.StatusOK)
			},
			jwt: validJWTWithoutUserAndPartner,
			buildRequest: func(jwt string) *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Add(auth.JWTCookie, fmt.Sprintf("%s %s", auth.BearerHeader, jwt))
				return r
			},
			statusCode: http.StatusOK,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			target.CommonDecodeHandler(tt.next)(recorder, tt.buildRequest(tt.jwt))
			require.Equal(t, recorder.Code, tt.statusCode)
		})
	}
}
