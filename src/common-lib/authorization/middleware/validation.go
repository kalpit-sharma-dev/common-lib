package middleware

import (
	"context"
	"crypto/md5"
	"errors"
	"net/http"

	"gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/auth"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token/signature"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/freecache"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

//go:generate mockgen -destination mock/mock.go -package mock . Responder,Cache

const (
	// in seconds
	jwtCacheTTL = 5 * 60
	// in bytes, default = 10MB
	defaultCacheSize = 10 * 1024 * 1024
)

var (
	// ErrNoPartnerID as a prerequisite partnerID should be specified in the request context
	ErrNoPartnerID = errors.New("no partnerID specified in the request.Context")
	// ErrNoUserID as a prerequisite userID should be specified in the request context
	ErrNoUserID = errors.New("no userID specified in the request.Context")
	// ErrBearerTokenInvalid indicates that JWT token is invalid
	ErrBearerTokenInvalid = errors.New("bearer token is missing, expired or invalid")
	// ErrAccessForbidden indicates that user access by provided JWT token is forbidden
	ErrAccessForbidden = errors.New("access is forbidden")
	// ErrNoSessionToken indicates that session token is not present
	ErrNoSessionToken = errors.New("no session token specified in the request.Context")
)

// TokenValidation is a JWT validation middleware
type TokenValidation struct {
	tExc      token.Exchanger
	tVal      signature.Validator
	responder Responder
	cache     Cache
	log       logger.Log
	trusting  bool
}

// Cache is an interface for cache
type Cache interface {
	Set(key, value []byte, ttl int) error
	Get(key []byte) ([]byte, error)
}

// Responder is an interface to return http response back to the caller.
// Usually it should be implemented by the user of the middleware to include their own custom localisation and build custom error message if required.
type Responder interface {
	Respond(ctx context.Context, resp Response, r *http.Request, rw http.ResponseWriter)
}

// defaultResponder is a default implementation of a responder that writes an error message to the body with given status code
type defaultResponder struct {
	log logger.Log
}

// Respond writes provided error message and status code to the response header and body
func (d *defaultResponder) Respond(ctx context.Context, resp Response, _ *http.Request, rw http.ResponseWriter) {
	rw.WriteHeader(resp.StatusCode)
	_, err := rw.Write([]byte(resp.Error.Error()))
	if err != nil {
		trID, ok := ctx.Value(auth.TransactionKey).(string)
		if !ok {
			trID = ""
		}

		d.log.Warn(trID, "failed to write response body: %s", err.Error())
	}
}

// Response is a container that stores information about error and http status code which should be returned to the caller
type Response struct {
	Error      error
	StatusCode int
}

// NewTokenValidation is a middleware that handles received JWT signature validation or
// in case it is not provided acquires JWT from the authorization service
func NewTokenValidation(authConfig token.AuthorizationConfig, exchanger token.Exchanger, validator signature.Validator,
	responder Responder, cache Cache, log logger.Log) *TokenValidation {

	if cache == nil {
		cache = freecache.New(defaultCacheSize)
	}

	if responder == nil {
		responder = &defaultResponder{
			log: log,
		}
	}

	if exchanger == nil {
		exchanger = token.NewJWTExchanger(authConfig, log)
	}

	if validator == nil {
		validator = signature.NewValidator(authConfig, log)
	}

	return &TokenValidation{
		tExc:      exchanger,
		tVal:      validator,
		responder: responder,
		cache:     cache,
		log:       log,
	}
}

// NewTrustingTokenValidation SHOULD ONLY BE USED IN SERVICES THAT CANNOT BE ACCESSED WITHOUT FIRST GOING THROUGH THE REVERSE PROXY OR OTHER SAFE GATEWAYS!
// It's like NewTokenValidation, but it won't validate the user ID or partner ID, and won't use a JWTExchanger (because the reverse proxy should have handled this already!)
func NewTrustingTokenValidation(authConfig token.AuthorizationConfig, validator signature.Validator, responder Responder, cache Cache, log logger.Log) *TokenValidation {
	if validator == nil {
		validator = signature.NewTrustingValidator(authConfig, log)
	}
	v := NewTokenValidation(authConfig, nil, validator, responder, cache, log)
	v.tExc = nil
	v.trusting = true
	return v
}

// Handler is an actual Validation middleware implementation, that validates received jwt or
// acquires it from the Authorization Service and then caches by provided cache implementation,
// by default inside in-memory freecache. It will also set auth.InternalRequest on the context.
//
// !!! This handler has prerequisites w/o which it cannot be used!
// One option is for you to use trusting mode (YOUR APP MUST NOT BE DIRECTLY ACCESSIBLE!).
// Trusting mode is appropriate if people can only access your app via the reverse proxy or a similarly secure gateway that handles things like partner ID checks.
// If you are not using trusting mode, then you must meet both of the following criteria:
// 1. Session token should be present in the request context, to use it for JWT exchange
// 2. PartnerID and UserID should be present in the context, there are a number of reasons why it should be there:
//   - for JWT claims validation in case it is passed in the cookies
//   - for JWT token exchange in Authorization MS
//
// Unless all requirements are met the handler will respond with `http.StatusInternalServerError` to the caller.
func (m *TokenValidation) Handler(next http.HandlerFunc) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		var accessToken, trID, partnerID, userID string
		var ok bool

		trID, ok = getStrFromCtx(r.Context(), auth.TransactionKey)
		if !ok {
			trID = ""
		}

		partnerID, ok = getStrFromCtx(r.Context(), auth.PartnerIDKey)
		if !ok && !m.trusting {
			m.responder.Respond(r.Context(), Response{
				Error:      ErrNoPartnerID,
				StatusCode: http.StatusInternalServerError,
			}, r, rw)
			return
		}

		userID, ok = getStrFromCtx(r.Context(), auth.UserIDKey)
		if !ok && !m.trusting {
			m.responder.Respond(r.Context(), Response{
				Error:      ErrNoUserID,
				StatusCode: http.StatusInternalServerError,
			}, r, rw)
			return
		}

		jwtToken, err := extractToken(r.Header)
		if err == nil {
			accessToken = jwtToken
			err = m.tVal.Validate(r.Context(), accessToken, partnerID, userID)
			if err != nil {
				status := http.StatusForbidden
				if err == token.ErrInvalidAuthzResponse {
					status = http.StatusInternalServerError
				}

				m.responder.Respond(r.Context(), Response{
					Error:      err,
					StatusCode: status,
				}, r, rw)
				return
			}
		} else if m.trusting {
			// If this is a trusting request and extractToken failed, then the only thing to check is if it's an internal request.
			// No exchange should happen, because for trusting, a gateway should have already done the exchange if the request isn't an internal request
			if !isInternalRequest(r.Header) {
				m.responder.Respond(r.Context(), Response{
					Error:      err,
					StatusCode: http.StatusUnauthorized,
				}, r, rw)
				return
			}
			// If this is an internal request, it will skip past the else block (exchanging tokens isn't a possibility, there is no token on an internal request)
		} else {
			sessionToken, ok := r.Context().Value(auth.SessionTokenKey).(string)
			if !ok {
				m.responder.Respond(r.Context(), Response{
					Error:      ErrNoSessionToken,
					StatusCode: http.StatusInternalServerError,
				}, r, rw)
				return
			}

			keyBytes := md5.Sum([]byte(sessionToken))
			cacheKey := keyBytes[:]

			tokenBytes, err := m.cache.Get(cacheKey)
			if err == nil {
				accessToken = string(tokenBytes)
			} else {
				accessToken, err = m.tExc.Exchange(r.Context(), sessionToken, partnerID)
				if err != nil {
					m.responder.Respond(r.Context(), Response{
						Error:      err,
						StatusCode: http.StatusInternalServerError,
					}, r, rw)
					return
				}

				err = m.cache.Set(cacheKey, []byte(accessToken), jwtCacheTTL)
				if err != nil {
					m.log.Warn(trID, "failed to set jwt to cache: %v", err)
				}
			}
		}

		ctx := context.WithValue(r.Context(), auth.JWTKey, accessToken)
		ctx = context.WithValue(ctx, auth.InternalRequest, isInternalRequest(r.Header))
		next(rw, r.WithContext(ctx))
	}
}

// Middleware is meant to be passed directly into the common lib router's Use method
func (m *TokenValidation) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(m.Handler(next.ServeHTTP))
}

func getStrFromCtx(ctx context.Context, key interface{}) (val string, found bool) {
	valFromCtx := ctx.Value(key)
	if valFromCtx == nil {
		return
	}
	val, found = valFromCtx.(string)
	return
}
