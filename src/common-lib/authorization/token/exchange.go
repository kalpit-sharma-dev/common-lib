package token

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/eapache/go-resiliency/retrier"
	"gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/auth"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/webClient"
)

const (
	DefaultHttpClientTimeout      = 1
	exchangeJWTFmtURL             = "%s/mgmt/partners/%s/users/token"
	exchangeEnhancedJWTURL        = "%s/mgmt/partners/%s/users/enhancedToken"
	serviceTokenExchangeJWTFmtURL = "%s/mgmt/assumeRole"

	partnerIDHeader = "partner_id"
	userIDHeader    = "user_id"
)

var (
	// ErrInvalidAuthzResponse is thrown on the invalid response from authorization service
	ErrInvalidAuthzResponse = errors.New("invalid response from authorization service")
	// Err5xxAuthzResponse is thrown on the 5xx response from authorization service
	Err5xxAuthzResponse = errors.New("5xx response from authorization service")
)

// Exchanger is an interface that handles JWT token acquisition by given provided session token
type Exchanger interface {
	Exchange(ctx context.Context, sessionToken, partnerID string) (string, error)
	ExchangeEnhanced(ctx context.Context, sessionToken, partnerID string) (string, error)
}

//go:generate mockgen -destination=mock/exchanger.go -package=mock . Exchanger

// JWTExchanger is a default implementation of `Exchanger`
type JWTExchanger struct {
	authorizationURL string
	httpClient       webClient.ClientService
	log              logger.Log
	rt               *retrier.Retrier
}

// AuthorizationConfig is a configuration that describes how calls to the authorization service should be made
type AuthorizationConfig struct {
	URL     string
	Client  webClient.ClientService
	Retrier *retrier.Retrier
}

// AuthClaims includes standard and custom claims specific to kksharmadev JWT implementation
type AuthClaims struct {
	jwt.StandardClaims
	CustomClaims
}

// CustomClaims includes claims specific to kksharmadev JWT implementation
type CustomClaims struct {
	UserID      string   `json:"user_id"`
	PartnerID   string   `json:"partner_id"`
	Permissions []string `json:"permissions"`
	// IsNoc indicates whether the user type is NOC or a Partner User, available as a part of the enhanced JWT.
	IsNoc *bool `json:"is_noc,omitempty"`
}

// NewJWTExchanger initializes default implementation of the `Exchanger`
func NewJWTExchanger(cfg AuthorizationConfig, log logger.Log) *JWTExchanger {
	if cfg.Client == nil {
		cfg.Client = DefaultWebClient()
	}

	if cfg.Retrier == nil {
		cfg.Retrier = DefaultRetrier()
	}

	return &JWTExchanger{
		authorizationURL: cfg.URL,
		httpClient:       cfg.Client,
		log:              log,
		rt:               cfg.Retrier,
	}
}

func (je *JWTExchanger) getToken(ctx context.Context, sessionToken, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	trID, ok := ctx.Value(auth.TransactionKey).(string)
	if !ok {
		trID = ""
	}

	req.Header.Add(auth.TransactionHeader, trID)
	req.Header.Add(auth.SessionTokenHeader, sessionToken)

	var resp *http.Response
	err = je.rt.Run(func() error {
		resp, err = je.httpClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode >= http.StatusInternalServerError {
			return Err5xxAuthzResponse
		}

		return nil
	})

	if err != nil && !errors.Is(err, Err5xxAuthzResponse) {
		return "", err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			je.log.Warn(trID, "failed to close: %s", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", ErrInvalidAuthzResponse
	}

	var response Token
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.Value, nil
}

// Exchange makes a call to Authorization service to acquire authorization JWT token by provided session token
func (je *JWTExchanger) Exchange(ctx context.Context, sessionToken, partnerID string) (string, error) {
	url := fmt.Sprintf(exchangeJWTFmtURL, je.authorizationURL, partnerID)
	return je.getToken(ctx, sessionToken, url)
}

// ExchangeEnhanced makes a call to Authorization service to acquire enhanced authorization JWT token by provided session token
func (je *JWTExchanger) ExchangeEnhanced(ctx context.Context, sessionToken, partnerID string) (string, error) {
	url := fmt.Sprintf(exchangeEnhancedJWTURL, je.authorizationURL, partnerID)
	return je.getToken(ctx, sessionToken, url)
}

// AssumeRoleRequest contains parameters for obtaining JWT token for server-2-server or user-2-server (if user
// specified) communication.
type AssumeRoleRequest struct {
	// RoleName a prototype role.
	RoleName string

	// PartnerID used for retrieving permissions against the user, user ID is also mandatory for this
	PartnerID string

	// UserID used for retrieving permissions against the user, partner ID is also mandatory for this
	UserID string
}

// AssumeRole makes a call to Authorization service to acquire authorization JWT token using prototype role.
// For user authorization both user ID and partner ID are mandatory. Providing only one or none will fall back to service token.
//
// The service name must be provided using env variable name from utils.ServiceNameEnv or service binary name would be used instead.
func (je *JWTExchanger) AssumeRole(ctx context.Context, assumeRoleRequest AssumeRoleRequest) (string, error) {
	url := fmt.Sprintf(serviceTokenExchangeJWTFmtURL, je.authorizationURL)

	serviceName := utils.GetServiceName()
	body, err := json.Marshal(map[string]string{
		"audience":  serviceName,
		"role_name": assumeRoleRequest.RoleName,
	})
	if err != nil {
		return "", fmt.Errorf("marshal servicename=%q and rolename=%q: %w", serviceName, assumeRoleRequest.RoleName, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("auth service request preparation: %w", err)
	}

	trID, ok := ctx.Value(auth.TransactionKey).(string)
	if !ok {
		trID = ""
	}
	req.Header.Add(auth.TransactionHeader, trID)

	if assumeRoleRequest.UserID != "" && assumeRoleRequest.PartnerID != "" {
		req.Header.Add(userIDHeader, assumeRoleRequest.UserID)
		req.Header.Add(partnerIDHeader, assumeRoleRequest.PartnerID)
	}

	var resp *http.Response
	err = je.rt.Run(func() error {
		resp, err = je.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("get JWT token: %w", err)
		}

		if resp.StatusCode >= http.StatusInternalServerError {
			return Err5xxAuthzResponse
		}

		return nil
	})

	if err != nil && !errors.Is(err, Err5xxAuthzResponse) {
		return "", err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			je.log.Warn(trID, "failed to close: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", ErrInvalidAuthzResponse
	}

	var response Token
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("decoding Auth response: %w", err)
	}

	return response.Value, nil
}

// Token is authorization service /token call response that returns a JWT token
type Token struct {
	Value string `json:"token"`
}

// DefaultWebClient returns default web client that will be used in case component's consumer will not provide their own
// to leverage circuit breaker, it should be registered first as shown in the ../webClient/example/example.go
func DefaultWebClient() webClient.ClientService {
	return webClient.ClientFactoryImpl{}.GetClientServiceByType(webClient.TLSClient,
		webClient.ClientConfig{TimeoutMinute: DefaultHttpClientTimeout})
}

// DefaultRetrier default implementation to retry http requests
func DefaultRetrier() *retrier.Retrier {
	const (
		tries = 3
		wait  = 5 * time.Millisecond
	)

	return retrier.New(retrier.ConstantBackoff(tries, wait), nil)
}
