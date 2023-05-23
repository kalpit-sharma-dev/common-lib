package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"time"

	cweb "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
)

// RequestContextKey Type to handle request context keys
type requestContextKey string

const (
	//RequestContextMappingKey is a key use in context request to store the value
	RequestContextMappingKey requestContextKey = "RequestContextMappingKey"
)

// AuthMiddleware is a middleware it will do some pre processing of the request
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxWithUser := context.WithValue(r.Context(), RequestContextMappingKey, "endpointmap")
		next.ServeHTTP(w, r.WithContext(ctxWithUser))
	})
}

// HandlerFuncWhenMiddleware will calll when middleware execution will complet
func HandlerFuncWhenMiddleware(w http.ResponseWriter, r *http.Request) {
	mapping := r.Context().Value(RequestContextMappingKey)
	fmt.Printf("Value from context%+v", mapping)
}

// HandlerFuncWhenwithoutmiddlware it will call when request arrived
func HandlerFuncWhenwithoutmiddlware(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handlerFuncWhenwithoutmiddlware invoked")
}

// Registering route with middleware
func addRouteWithMiddleware(r1 cweb.Router) {
	r1.AddHandle("/path-to-api", AuthMiddleware(HandlerFuncWhenMiddleware), http.MethodGet)
}

// Registering route without middleware
func addRouteWithoutMiddleware(r2 cweb.Router) {
	r2.AddFunc("/apth-to-api", HandlerFuncWhenwithoutmiddlware, http.MethodGet)
}

// Registering PPROF routes.
func addPPROFRoutes(router cweb.Router) {
	router.AddFunc("/debug/pprof/", pprof.Index)
	router.AddFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.AddFunc("/debug/pprof/profile", pprof.Profile)
	router.AddFunc("/debug/pprof/symbol", pprof.Symbol)
	router.AddFunc("/debug/pprof/trace", pprof.Trace)

	// Manually add support for paths linked to by index page at /debug/pprof/
	router.AddHandle("/debug/pprof/allocs", pprof.Handler("allocs"))
	router.AddHandle("/debug/pprof/block", pprof.Handler("block"))
	router.AddHandle("/debug/pprof/cmdline", pprof.Handler("cmdline"))
	router.AddHandle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.AddHandle("/debug/pprof/heap", pprof.Handler("heap"))
	router.AddHandle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.AddHandle("/debug/pprof/mutex", pprof.Handler("mutex"))
}

type testMiddleware struct {
	next http.Handler
}

func (h *testMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Middleware function invoked")
	content := []byte("test")
	w.Write(content)
	h.next.ServeHTTP(w, r)
}

func (h *testMiddleware) dummyMiddleware(handler http.Handler) http.Handler {
	return &testMiddleware{next: handler}
}

// Add middleware to router
func addMiddlewareToRouter(r cweb.Router) {
	tm := &testMiddleware{}
	//add variadic middleware
	r.Use(tm.dummyMiddleware, tm.dummyMiddleware, tm.dummyMiddleware)
}

// getting Server Config Object
var getServerConfig = func() *cweb.ServerConfig {
	return &cweb.ServerConfig{ListenURL: ":8080"}
}

// getObservabilityServerConfig gets observability Server Config.
func getObservabilityServerConfig() *cweb.ServerConfig {
	return &cweb.ServerConfig{ListenURL: ":8081"}
}

// getting Server Object
var createServer = func(cfg *cweb.ServerConfig) cweb.HTTPServer {
	return cweb.Create(cfg)
}

func main() {
	ctx, h := setup()
	runPPROF(ctx)
	log.Fatal(h.Start(ctx))
}

func setup() (context.Context, cweb.HTTPServer) {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)

	// getting Server Object
	cfg := getServerConfig()
	// getting Server Config Object
	h := createServer(cfg)
	// getting associated router
	r1 := h.GetRouter()
	addRouteWithMiddleware(r1)
	addRouteWithoutMiddleware(r1)
	addMiddlewareToRouter(r1)
	return ctx, h
}

func runPPROF(ctx context.Context) {
	// registering PPROF server on different address:
	pprofCfg := getObservabilityServerConfig()
	pprofSrv := createServer(pprofCfg)
	pprofRouter := pprofSrv.GetRouter()
	addPPROFRoutes(pprofRouter)
	go func() {
		log.Printf("serving PPROF on %v", pprofCfg.ListenURL)
		log.Fatal(pprofSrv.Start(ctx))
	}()
}
