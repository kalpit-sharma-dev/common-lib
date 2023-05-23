package web

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/http2"

	"github.com/gorilla/mux"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/tracing"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

const (
	cProxyHeader = "X-FORWARDED-FOR"
)

// muxConfig structure is Mux Adapter for Server Interface
type muxConfig struct {
	serverCfg  *ServerConfig
	router     *mux.Router
	srv        *http.Server
	tracingCfg *tracing.Config
}

// SetupRoutes implementation of Server interface for muxConfig
func (mcfg *muxConfig) SetupRoutes(routes []*RouteConfig) {
	for _, route := range routes {
		handler := muxRouteHandler{route: route}
		mcfg.router.HandleFunc(mcfg.serverCfg.URLPathPrefix+route.URLPathSuffix, handler.handleFunc)
	}

	staticFileDirectory := mcfg.serverCfg.StaticFileDirectory
	if strings.TrimSpace(staticFileDirectory) != "" {
		path := rest.FilePath(mcfg.serverCfg.URLPathPrefix)
		f := rest.FileHandler(staticFileDirectory, path)
		mcfg.router.PathPrefix(path).Handler(f)
	}
}

// SetRouter sets the new mux router to the http handler
func (mcfg *muxConfig) SetRouter(r *mux.Router) {
	mcfg.srv.Handler = r
}

// GetRouter returns the mux router
func (mcfg *muxConfig) GetRouter() *mux.Router {
	return mcfg.router
}

// Convert converts the RouteConfig to mux router
func (mcfg *muxConfig) Convert(routes []*RouteConfig) *mux.Router {
	r := mux.NewRouter()
	for _, route := range routes {
		handler := muxRouteHandler{route: route}
		r.HandleFunc(mcfg.serverCfg.URLPathPrefix+route.URLPathSuffix, handler.handleFunc)
	}
	return r
}

// Callback hook for Graceful shutdown capabillities for the server
func (mcfg *muxConfig) RegisterOnShutdown(f func()) {
	mcfg.srv.RegisterOnShutdown(f)
}

// CreateServer creates http server
func (mcfg *muxConfig) CreateServer() {
	mcfg.createServerInstance()
}

// Shutdown gracefully shutsdown http server
func (mcfg *muxConfig) ShutDown(ctx context.Context) error {
	if mcfg.srv != nil {
		return mcfg.srv.Shutdown(ctx)
	}
	return nil
}

// ListenAndServe implementation of Server interface for muxConfig
func (mcfg *muxConfig) ListenAndServe() error {
	handler, err := tracing.WrapHandlerWithTracing(mcfg.tracingCfg, mcfg.router)
	if err != nil {
		handler = mcfg.router
	}
	mcfg.srv.Handler = handler
	return mcfg.srv.ListenAndServe()
}

// HTTP2ListenAndServeTLS listen as HTTP2 (which is only in TLS)
func (mcfg *muxConfig) HTTP2ListenAndServeTLS() error {
	mcfg.createServerInstance()
	err := http2.ConfigureServer(mcfg.srv, &http2.Server{
		IdleTimeout:          time.Duration(mcfg.serverCfg.IdleTimeoutMinute) * time.Minute,
		MaxHandlers:          mcfg.serverCfg.MaxHandlers,
		MaxConcurrentStreams: mcfg.serverCfg.MaxConcurrentStreams,
	})

	if err != nil {
		return err
	}

	return mcfg.srv.ListenAndServeTLS(mcfg.serverCfg.CertificateFile, mcfg.serverCfg.CertificateKeyFile)
}

func (mcfg *muxConfig) createServerInstance() {
	if mcfg.srv == nil {
		handler, err := tracing.WrapHandlerWithTracing(mcfg.tracingCfg, mcfg.router)
		if err != nil {
			handler = mcfg.router
		}
		mcfg.srv = &http.Server{
			Addr:         mcfg.serverCfg.ListenURL,
			Handler:      handler,
			ReadTimeout:  time.Duration(mcfg.serverCfg.ReadTimeoutMinute) * time.Minute,
			WriteTimeout: time.Duration(mcfg.serverCfg.WriteTimeoutMinute) * time.Minute,
		}

	}
}

type muxRouteHandler struct {
	route *RouteConfig
}

type muxRequestContext struct {
	response      http.ResponseWriter
	request       *http.Request
	vars          map[string]string
	varsResolved  bool
	dcDateTimeUTC time.Time
}

func (ctx muxRequestContext) GetRequest() *http.Request {
	return ctx.request
}

func (ctx muxRequestContext) GetResponse() http.ResponseWriter {
	return ctx.response
}

func (ctx muxRequestContext) GetVars() map[string]string {
	if !ctx.varsResolved {
		ctx.varsResolved = true
		ctx.vars = mux.Vars(ctx.request)
	}
	return ctx.vars
}

func (ctx muxRequestContext) GetData() (data []byte, err error) {
	return ioutil.ReadAll(ctx.GetRequest().Body)
}

func (ctx muxRequestContext) GetRequestDcDateTimeUTC() time.Time {
	return ctx.dcDateTimeUTC
}

func (aHandler muxRouteHandler) handleFunc(w http.ResponseWriter, r *http.Request) {
	ctx := &muxRequestContext{
		response:      w,
		request:       r,
		dcDateTimeUTC: time.Now().UTC(),
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch r.Method {
	case http.MethodGet:
		aHandler.route.Res.Get(ctx)
	case http.MethodPost:
		aHandler.route.Res.Post(ctx)
	case http.MethodPut:
		aHandler.route.Res.Put(ctx)
	case http.MethodDelete:
		aHandler.route.Res.Delete(ctx)
	default:
		aHandler.route.Res.Others(ctx)
	}
}

func (ctx muxRequestContext) GetRemoteAddr() (string, error) {
	return GetRemoteAddress(ctx.GetRequest())
}

// GetRequestContext convert w,r to RequestContext
func GetRequestContext(w http.ResponseWriter, r *http.Request) RequestContext {
	return &muxRequestContext{
		response:      w,
		request:       r,
		dcDateTimeUTC: time.Now().UTC(),
	}
}

// GetRemoteAddress returns the remote address based on the request
var GetRemoteAddress = func(r *http.Request) (string, error) {
	ipAddress := "0.0.0.0"
	remoteProxy := r.Header.Get(cProxyHeader)
	remoteHostPort := r.RemoteAddr
	//If remoteProxy is set it means endpoint was hidden behind proxy, hence get the real IP from x-forwarded-for header.
	//else try to get it directly from RemoteAddress attribute of HTTP request
	if len(remoteProxy) > 0 {
		// X-Forwarded-For : client, proxy1, proxy2
		// where the value is a comma+space separated list of IP addresses
		ips := strings.Split(remoteProxy, ", ")
		if len(ips) > 0 {
			//TODO verify if first IP is the real IP.
			ip := net.ParseIP(ips[0])
			if ip != nil {
				ipAddress = ip.String()
			}
		}
	} else if len(remoteHostPort) > 0 {
		host, _, err := net.SplitHostPort(remoteHostPort)
		if err != nil {
			return ipAddress, err
		}
		ip := net.ParseIP(host)
		if ip != nil {
			ipAddress = ip.String()
		}
	}
	return ipAddress, nil
}
