package web

import (
	"context"
	"net/http"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/tracing"

	"github.com/gorilla/mux"
)

//go:generate mockgen -package mock -destination=mock/mocks.go . Server,ServerFactory,Resource,RequestContext,HTTPServer,Router

// ServerConfig stores the configuration of the web server.
type ServerConfig struct {
	URLPathPrefix        string
	ListenURL            string
	CertificateFile      string
	CertificateKeyFile   string
	ReadTimeoutMinute    int
	WriteTimeoutMinute   int
	IdleTimeoutMinute    int
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	MaxHandlers          int
	MaxConcurrentStreams uint32
	StaticFileDirectory  string
	APIversion           string
	TracingConfig        *tracing.Config
}

// Server interface sets up routes, handlers and listening.
type Server interface {
	SetupRoutes(res []*RouteConfig)
	ListenAndServe() error
	RegisterOnShutdown(f func())
	ShutDown(ctx context.Context) error
	GetRouter() *mux.Router
	Convert(routes []*RouteConfig) *mux.Router
	SetRouter(r *mux.Router)
	HTTP2ListenAndServeTLS() error
	CreateServer()
}

// RequestContext that will be provided by router to request handlers
type RequestContext interface {
	GetResponse() http.ResponseWriter
	GetRequest() *http.Request
	GetVars() map[string]string
	GetRequestDcDateTimeUTC() time.Time
	GetData() (data []byte, err error)
	GetRemoteAddr() (string, error)
}

// ServerFactory interface to for a Factory impmementation
type ServerFactory interface {
	GetServer(cfg *ServerConfig) Server
}

// ServerFactoryImpl A factory implementation for the HTTP server creation
type ServerFactoryImpl struct{}

// GetServer implements Server interface
func (ServerFactoryImpl) GetServer(cfg *ServerConfig) Server {
	mcfg := muxConfig{
		serverCfg: cfg,
		router:    mux.NewRouter(),
		srv: &http.Server{
			Addr:         cfg.ListenURL,
			ReadTimeout:  time.Duration(cfg.ReadTimeoutMinute) * time.Minute,
			WriteTimeout: time.Duration(cfg.WriteTimeoutMinute) * time.Minute,
		},
		tracingCfg: cfg.TracingConfig,
	}
	return &mcfg
}
