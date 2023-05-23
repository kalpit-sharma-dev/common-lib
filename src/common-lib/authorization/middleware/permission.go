package middleware

import (
	"context"
	"net/http"
	"strings"

	"gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/auth"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token/permission"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
)

// Permission is a middleware to validate user permissions against a route
type Permission struct {
	rp       Responder
	log      logger.Log
	trusting bool
}

// ErrorCodes for Logging
const ErrCodeBearerTokenInvalid = "Permission.ErrBearerTokenInvalid"
const ErrCodeAccessForbidden = "Permission.ErrAccessForbidden"

// NewPermission initializes permission middleware
func NewPermission(rp Responder, log logger.Log) *Permission {
	if rp == nil {
		rp = &defaultResponder{log}
	}

	return &Permission{
		rp:  rp,
		log: log,
	}
}

// NewTrustingPermission SHOULD ONLY BE USED IN SERVICES THAT CANNOT BE ACCESSED WITHOUT FIRST GOING THROUGH THE REVERSE PROXY OR OTHER SAFE GATEWAYS!
// It's like NewPermission, but if there is no JWT, then it assumes the user is allowed
func NewTrustingPermission(rp Responder, log logger.Log) *Permission {
	p := NewPermission(rp, log)
	p.trusting = true
	return p
}

// If true, then allow the request no matter what. Don't check the JWT or do anything with it
func (p *Permission) trustUnconditionally(headers http.Header) bool {
	return p.trusting && isInternalRequest(headers)
}

// CommonAssertHandler wraps a common lib handler with middleware that will validate user permissions based on PermissionValidationType
func (p *Permission) CommonAssertHandler(handler web.HTTPHandlerFunc, validationType permission.ValidationType, permissionNames ...string) web.HTTPHandlerFunc {
	return web.HTTPHandlerFunc(p.AssertHandler(http.HandlerFunc(handler), permissionNames, validationType))
}

// CommonDecodeHandler wraps a common lib handler with a middleware that decodes a token permissions and sets it to the request context
func (p *Permission) CommonDecodeHandler(handler web.HTTPHandlerFunc) web.HTTPHandlerFunc {
	return web.HTTPHandlerFunc(p.DecodeHandler(http.HandlerFunc(handler)))
}

// AssertHandler wraps a handler with a middleware that will validate user permissions based on PermissionValidationType
func (p *Permission) AssertHandler(next http.HandlerFunc, permissions []string,
	validationType permission.ValidationType) func(rw http.ResponseWriter, r *http.Request) {

	return func(rw http.ResponseWriter, r *http.Request) {
		if p.trustUnconditionally(r.Header) {
			next(rw, r)
			return
		}

		var isAllowed bool
		// If DecodeHandler/CommonDecodeHandler have already run, then reuse what they decoded into context rather, than trying to decode the JWT again
		if perms := r.Context().Value(auth.PermissionKey); perms != nil {
			isAllowed = permission.ValidatePermissions(perms.([]string), permissions, validationType)
		} else {
			jwt, err := extractToken(r.Header)
			if err != nil {
				p.log.ErrorC(r.Context(), ErrCodeBearerTokenInvalid, "Error in Extracting JWT token from Headers")
				p.rp.Respond(r.Context(), Response{
					Error:      err,
					StatusCode: http.StatusUnauthorized,
				}, r, rw)
				return
			}

			isAllowed, err = permission.Validate(jwt, permissions, validationType)
			if err != nil {
				p.log.ErrorC(r.Context(), ErrCodeBearerTokenInvalid, "Error in Decoding Authorization JWT token")
				p.rp.Respond(r.Context(), Response{
					Error:      ErrBearerTokenInvalid,
					StatusCode: http.StatusUnauthorized,
				}, r, rw)
				return
			}
		}

		if !isAllowed {
			p.log.ErrorC(r.Context(), ErrCodeAccessForbidden, "JWT token provided does not have the necessary permissions")
			p.rp.Respond(r.Context(), Response{
				Error:      ErrAccessForbidden,
				StatusCode: http.StatusForbidden,
			}, r, rw)
			return
		}

		next(rw, r)
	}
}

// DecodeHandler wraps a handler with a middleware that decodes a token permissions and sets it to the request context
func (p *Permission) DecodeHandler(next http.HandlerFunc) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		if p.trustUnconditionally(r.Header) {
			next(rw, r)
			return
		}

		jwt, err := extractToken(r.Header)
		if err != nil {
			p.log.ErrorC(r.Context(), ErrCodeBearerTokenInvalid, "Error in Extracting JWT token from Headers")
			p.rp.Respond(r.Context(), Response{
				Error:      err,
				StatusCode: http.StatusUnauthorized,
			}, r, rw)
			return
		}

		claims, err := permission.DecodeClaims(jwt)
		if err != nil {
			p.log.ErrorC(r.Context(), ErrCodeBearerTokenInvalid, "Error in Decoding Authorization JWT token")
			p.rp.Respond(r.Context(), Response{
				Error:      ErrBearerTokenInvalid,
				StatusCode: http.StatusUnauthorized,
			}, r, rw)
			return
		}

		ctx := context.WithValue(r.Context(), auth.PermissionKey, claims.Permissions)
		// "Assume Role" feature will issue token without user and partner context:
		if claims.UserID != "" {
			ctx = context.WithValue(ctx, auth.UserIDKey, claims.UserID)
		}
		if claims.PartnerID != "" {
			ctx = context.WithValue(ctx, auth.PartnerIDKey, claims.PartnerID)
		}
		ctx = context.WithValue(ctx, auth.JWTKey, jwt)

		next(rw, r.WithContext(ctx))
	}
}

func extractToken(header http.Header) (string, error) {
	tkn := strings.TrimSpace(strings.TrimPrefix(header.Get(auth.AuthorizationHeader), auth.BearerHeader))
	if tkn == "" {
		return "", ErrBearerTokenInvalid
	}

	return tkn, nil
}

func isInternalRequest(header http.Header) bool {
	return header.Get(auth.AuthorizationHeader) == ""
}
