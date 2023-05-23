// Package webClient package abstracts the underlying http packages used;
// this would help abstract cross cutting concerns like Encryption, Compression etc
package webClient

import (
	"net/http"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/tracing"
)

//go:generate mockgen -package mock -destination=mock/mocks.go . ClientFactory,ClientService,HTTPClientFactory,HTTPClientService

// ClientFactory provides implementation of ClientService
type ClientFactory interface {
	GetClientService(f HTTPClientFactory, config ClientConfig) ClientService
	GetClientServiceByType(clientType ClientType, config ClientConfig) HTTPClientService
}

// ClientService to be implemented by HTTP Web client
type ClientService interface {
	Do(request *http.Request) (*http.Response, error)
}

// HTTPClientFactory provides implementation of HttpCommandService
type HTTPClientFactory interface {
	GetHTTPClient(config ClientConfig) HTTPClientService
}

// HTTPClientService provides methods for posting data using the net/http package
type HTTPClientService interface {
	Create()
	Do(request *http.Request) (*http.Response, error)
	SetCheckRedirect(func(req *http.Request, via []*http.Request) error)
}

// ClientConfig Http Client Configuration for HTTP connection
type ClientConfig struct {
	MaxIdleConns                int
	MaxIdleConnsPerHost         int
	IdleConnTimeoutMinute       int
	TimeoutMinute               int
	TimeoutMillisecond          int
	DialTimeoutSecond           int
	DialKeepAliveSecond         int
	TLSHandshakeTimeoutSecond   int
	ExpectContinueTimeoutSecond int
	UseIEProxy                  bool
	ProxySetting                ProxySetting
	ValidateSSLCertificate      bool
	TracingConfig               *tracing.Config
}

// ProxySetting is the struct for Proxy settings
type ProxySetting struct {
	IP       string
	Port     int
	UserName string
	Password string
	Protocol string
}

// MessageType would specify whether message needs to be sent to
// Broker or some other location, this way location of the server can be configured at
// single location
type MessageType uint32

const (
	// Broker Message Type
	Broker MessageType = 1
)

// HTTPMethod would specify the http method to be executed
type HTTPMethod uint32

const (
	// Post method of HTTP
	Post HTTPMethod = 1
)

// Error Codes
const (
	ErrorInvalidHTTPMethod  = "ErrInvalidHTTPMethod"
	ErrorEmptyContentType   = "ErrEmptyContentType"
	ErrorNilURL             = "ErrorNilURL"
	ErrorBlankHTTPMethod    = "BlankHttpMethod"
	ErrorNilData            = "ErrNilData"
	ErrorInvalidMessageType = "ErrInvalidMessageType"
	ErrorEmptyURLSuffix     = "ErrEmptyURLSuffix"
	ErrorInvalidURLSuffix   = "ErrInvalidURLSuffix"

	BasicClient ClientType = 10
	TLSClient   ClientType = 20
)

// ClientType client type
type ClientType int
