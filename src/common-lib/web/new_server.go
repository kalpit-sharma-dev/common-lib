package web

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

//HTTPServer interface start/stop the server.
type HTTPServer interface {
	//Start implementation of Server interface for newMuxConfig
	Start(ctx context.Context) error
	//Callback hook for Graceful shutdown capabillities for the server
	RegisterOnShutdown(f func())
	//Shutdown gracefully shutsdown http server
	ShutDown(ctx context.Context) error
	// GetRouter will return the Router
	GetRouter() Router
}

//Create get instance of HTTPServer
func Create(cfg *ServerConfig) HTTPServer {
	return &newMuxConfig{
		serverCfg: cfg,
		router:    &gorillaRouter{mux.NewRouter()},
		srv: &http.Server{
			Addr:         cfg.ListenURL,
			ReadTimeout:  time.Duration(cfg.ReadTimeoutMinute) * time.Minute,
			WriteTimeout: time.Duration(cfg.WriteTimeoutMinute) * time.Minute,
		},
		tracingCfg: cfg.TracingConfig,
	}
}
