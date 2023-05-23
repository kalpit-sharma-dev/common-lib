package web

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/tracing"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

const (
	serverDoesNotExist = "Server Does Not Exist"
)

// HTTPHandlerFunc is type which will be used when we are calling AddFunc method of Router interface, so means without middleware
type HTTPHandlerFunc func(w http.ResponseWriter, r *http.Request)

// serverRouter is an external interface of router interface
type serverRouter interface {
	Router
	//pathPrefix registers a new route with a matcher for the URL path prefix
	pathPrefix(path string) *mux.Route
}

// Router will registering a couple of URL paths and handlers
type Router interface {
	// AddFunc registers HTTPHandlerFunc as a route with optional methods.
	AddFunc(route string, handleFunc HTTPHandlerFunc, methods ...string)
	// AddHandle registers http.Handler as a route with optional methods.
	AddHandle(route string, handler http.Handler, methods ...string)
	// Use appends a middleware functions to the chain.
	// Middleware can be used to intercept or otherwise modify requests and/or responses, and are executed in the order that they are applied to the Router.
	Use(handleFunc ...func(http.Handler) http.Handler)
	// ServeHTTP is http handler function
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// GorillaRouter is wrapper of mux router
type gorillaRouter struct {
	router *mux.Router
}

// ServeHTTP method for handler
func (grouter *gorillaRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	grouter.router.ServeHTTP(w, r)
}

// AddFunc registers HTTPHandlerFunc as a route with optional methods.
func (grouter *gorillaRouter) AddFunc(route string, handleFunc HTTPHandlerFunc, methods ...string) {
	r := grouter.router.HandleFunc(route, handleFunc)
	// add a matchers for HTTP methods only if there are methods specified. Otherwise do not restrict route on per HTTP method basis:
	if len(methods) != 0 {
		r.Methods(methods...)
	}
}

// Use adds middleware to router
func (grouter *gorillaRouter) Use(handleFunc ...func(http.Handler) http.Handler) {
	for _, fn := range handleFunc {
		grouter.router.Use(fn)
	}
}

// AddHandle registers http.Handler as a route with optional methods.
func (grouter *gorillaRouter) AddHandle(route string, handler http.Handler, methods ...string) {
	r := grouter.router.Handle(route, handler)
	// add a matcher for HTTP methods only if there are methods specified. Otherwise do not restrict route on per HTTP method basis:
	if len(methods) != 0 {
		r.Methods(methods...)
	}
}

// pathPrefix registers a new route with a matcher for the URL path prefix
func (grouter *gorillaRouter) pathPrefix(path string) *mux.Route {
	return grouter.router.PathPrefix(path)
}

// newMuxConfig structure is Mux Adapter for Server Interface
type newMuxConfig struct {
	serverCfg  *ServerConfig
	router     serverRouter
	srv        *http.Server
	tracingCfg *tracing.Config
}

// Callback hook for Graceful shutdown capabillities for the server
func (mcfg *newMuxConfig) RegisterOnShutdown(f func()) {
	mcfg.srv.RegisterOnShutdown(f)
}

// Shutdown gracefully shutsdown http server
func (mcfg *newMuxConfig) ShutDown(ctx context.Context) error {
	if mcfg.srv != nil {
		return mcfg.srv.Shutdown(ctx)
	}
	return errors.New(serverDoesNotExist)
}

// Start implementation of Server interface for newMuxConfig
func (mcfg *newMuxConfig) Start(ctx context.Context) error {
	staticFileDirectory := mcfg.serverCfg.StaticFileDirectory
	if strings.TrimSpace(staticFileDirectory) != "" {
		path := rest.FilePath(mcfg.serverCfg.URLPathPrefix)
		path = path + mcfg.serverCfg.APIversion + "/download/"
		f := rest.FileHandler(staticFileDirectory, path)
		mcfg.router.pathPrefix(path).Handler(f)
	}
	var err error
	mcfg.srv.Handler, err = tracing.WrapHandlerWithTracing(mcfg.tracingCfg, mcfg.router)
	if err != nil {
		mcfg.srv.Handler = mcfg.router
	}
	return mcfg.srv.ListenAndServe()
}

// GetRouter will return the Router
func (mcfg *newMuxConfig) GetRouter() Router {
	return mcfg.router
}
