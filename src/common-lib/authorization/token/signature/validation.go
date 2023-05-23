package signature

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/dgrijalva/jwt-go"
	"github.com/eapache/go-resiliency/retrier"
	"gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/auth"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/authorization/token"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/webClient"
)

//go:generate mockgen -destination=../mock/validator.go -package=mock . Validator,TokenParser

const (
	kidJWTHeader              = "kid"
	authorizationPubKeyURLFmt = "%s/mgmt/keys/%s"
)

var errInvalidBearerToken = errors.New("bearer token is invalid")

type (
	// Validator is an interface for JWT validation
	Validator interface {
		Validate(ctx context.Context, jwtToken, partnerID, userID string) error
	}

	// TokenParser is an interface for bearer token parsing
	TokenParser interface {
		Parse(ctx context.Context, bearerToken string) (*token.CustomClaims, error)
	}

	// LocalSignatureValidator is a local validator of JWT signature and claims
	LocalSignatureValidator struct {
		pubKeys              sync.Map
		client               webClient.ClientService
		authorizationURL     string
		log                  logger.Log
		rt                   *retrier.Retrier
		optionalIDValidation bool
	}
)

// NewValidator returns initialized local signature validator implementation
func NewValidator(cfg token.AuthorizationConfig, log logger.Log) *LocalSignatureValidator {
	if cfg.Client == nil {
		cfg.Client = token.DefaultWebClient()
	}

	if cfg.Retrier == nil {
		cfg.Retrier = token.DefaultRetrier()
	}

	return &LocalSignatureValidator{
		client:           cfg.Client,
		authorizationURL: cfg.URL,
		log:              log,
		rt:               cfg.Retrier,
	}
}

// NewTrustingValidator SHOULD ONLY BE USED IN SERVICES THAT CANNOT BE ACCESSED WITHOUT FIRST GOING THROUGH THE REVERSE PROXY OR OTHER SAFE GATEWAYS!
// It's like NewValidator, but it won't validate the user ID or partner ID if they are empty
func NewTrustingValidator(cfg token.AuthorizationConfig, log logger.Log) *LocalSignatureValidator {
	v := NewValidator(cfg, log)
	v.optionalIDValidation = true
	return v
}

// Validate parses input JWT token, validates JWT signature and claims
func (sg *LocalSignatureValidator) Validate(ctx context.Context, jwtToken, partnerID, userID string) error {
	var claims token.AuthClaims
	_, err := jwt.ParseWithClaims(jwtToken, &claims, sg.parseToken(ctx))
	if err != nil {
		return err
	}

	return validateIdentity(claims, sg.optionalIDValidation, partnerID, userID)
}

// Parse parses input bearer token and returns its claims
func (sg *LocalSignatureValidator) Parse(ctx context.Context, bearerToken string) (*token.CustomClaims, error) {
	jwtToken := strings.TrimSpace(strings.TrimPrefix(bearerToken, auth.BearerHeader))
	if jwtToken == "" {
		return nil, errInvalidBearerToken
	}

	var claims token.AuthClaims
	_, err := jwt.ParseWithClaims(jwtToken, &claims, sg.parseToken(ctx))
	if err != nil {
		return nil, err
	}

	return &claims.CustomClaims, nil
}

func (sg *LocalSignatureValidator) parseToken(ctx context.Context) jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header[kidJWTHeader]
		if !ok {
			return nil, errors.New("JWT token does not contain kid header")
		}

		kidStr, ok := kid.(string)
		if !ok {
			return nil, errors.New("JWT token contains invalid kid")
		}

		pubKey, err := sg.getPubKey(ctx, kidStr)
		if err != nil {
			return nil, err
		}

		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return jwt.ParseRSAPublicKeyFromPEM([]byte(pubKey))
	}
}

// This will try to get the cached pubKey, and fall back to running fetchPubKey and storing its value in the cache
func (sg *LocalSignatureValidator) getPubKey(ctx context.Context, kid string) (string, error) {
	val, ok := sg.pubKeys.Load(kid)
	if ok {
		return val.(string), nil
	}

	// Ok, time to actually go to authz
	pubKey, err := sg.fetchPubKey(ctx, kid)
	if err == nil {
		sg.pubKeys.Store(kid, pubKey)
	}
	return pubKey, err
}

// Unlike getPubKey, this will go to authz 100% of the time
func (sg *LocalSignatureValidator) fetchPubKey(ctx context.Context, kid string) (string, error) {
	type authzRsp struct {
		Key string `json:"public_key"`
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(authorizationPubKeyURLFmt, sg.authorizationURL, kid), nil)
	if err != nil {
		return "", err
	}

	trID, ok := ctx.Value(auth.TransactionKey).(string)
	if !ok {
		trID = ""
	}
	req.Header.Add(auth.TransactionHeader, trID)

	var rsp *http.Response
	err = sg.rt.Run(func() error {
		rsp, err = sg.client.Do(req)
		if err != nil {
			return err
		}

		if rsp.StatusCode >= http.StatusInternalServerError {
			return token.Err5xxAuthzResponse
		}

		return nil
	})

	if err != nil && err != token.Err5xxAuthzResponse {
		return "", err
	}

	defer func() {
		if e := rsp.Body.Close(); e != nil {
			sg.log.Warn(trID, "failed to close: %s", e)
		}
	}()

	if (rsp == nil || rsp.Body == nil) || rsp.StatusCode != http.StatusOK {
		return "", token.ErrInvalidAuthzResponse
	}

	body := new(authzRsp)
	err = json.NewDecoder(rsp.Body).Decode(body)
	if err != nil {
		return "", err
	}

	return body.Key, nil
}

func validateIdentity(claims token.AuthClaims, skipIDValidation bool, partnerID, userID string) error {
	if claims.PartnerID != partnerID && !(partnerID == "" && skipIDValidation) {
		return errors.New("partner claim is invalid")
	}

	if userID != "" && claims.UserID != userID {
		return errors.New("user claim is invalid")
	}

	return nil
}
