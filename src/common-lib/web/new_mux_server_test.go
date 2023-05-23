package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGorillaRouter_AddFunc(t *testing.T) {
	var test HTTPHandlerFunc
	type fields struct {
		Router *mux.Router
	}
	type args struct {
		route      string
		handleFunc HTTPHandlerFunc
		methods    []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Test for execution of method without middleware",
			fields: fields{Router: mux.NewRouter()},
			args:   args{route: "/test", handleFunc: test, methods: []string{"DELETE"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grouter := &gorillaRouter{
				router: tt.fields.Router,
			}
			grouter.AddFunc(tt.args.route, tt.args.handleFunc, tt.args.methods...)
		})
	}
}

func TestGorillaRouter_AddHandle(t *testing.T) {
	type fields struct {
		router *mux.Router
	}
	type args struct {
		route   string
		handler http.Handler
		methods []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Test for execution of method with middleware",
			fields: fields{router: mux.NewRouter()},
			args:   args{route: "/test", handler: nil, methods: []string{"POST"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grouter := &gorillaRouter{
				router: tt.fields.router,
			}
			grouter.AddHandle(tt.args.route, tt.args.handler, tt.args.methods...)
		})
	}
}

func Test_newMuxConfig_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type fields struct {
		serverCfg *ServerConfig
		router    serverRouter
		srv       *http.Server
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "To test the start server",
			fields:  fields{serverCfg: &ServerConfig{}, router: &gorillaRouter{mux.NewRouter()}, srv: &http.Server{}},
			args:    args{ctx: context.Background()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcfg := &newMuxConfig{
				serverCfg: tt.fields.serverCfg,
				router:    tt.fields.router,
				srv:       tt.fields.srv,
			}
			quit := make(chan bool)
			go func() {
				for range time.Tick(10 * time.Millisecond) {
					//stop the server if it launched successfully so testing can continue
					//ErrServerClosed error returns from shutdown to pass the test
					select {
					case <-quit:
						return
					default:
						mcfg.ShutDown(tt.args.ctx)
					}
				}
			}()
			if err := mcfg.Start(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("newMuxConfig.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
			quit <- true
		})
	}
}

func Test_newMuxConfig_GetRouter(t *testing.T) {
	type fields struct {
		serverCfg *ServerConfig
		router    serverRouter
		srv       *http.Server
	}
	tests := []struct {
		name   string
		fields fields
		want   Router
	}{
		{
			name:   "To test the Get Router",
			fields: fields{serverCfg: &ServerConfig{}, router: &gorillaRouter{mux.NewRouter()}, srv: &http.Server{}},
			want:   &gorillaRouter{mux.NewRouter()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcfg := &newMuxConfig{
				serverCfg: tt.fields.serverCfg,
				router:    tt.fields.router,
				srv:       tt.fields.srv,
			}
			if got := mcfg.GetRouter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMuxConfig.GetRouter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newMuxConfig_ShutDown(t *testing.T) {
	type fields struct {
		serverCfg *ServerConfig
		router    serverRouter
		srv       *http.Server
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Shut Down Success",
			fields:  fields{serverCfg: &ServerConfig{}, router: &gorillaRouter{mux.NewRouter()}, srv: &http.Server{}},
			args:    args{ctx: context.Background()},
			wantErr: false,
		},
		{
			name:    "Shut Down Fail",
			fields:  fields{serverCfg: &ServerConfig{}, router: &gorillaRouter{mux.NewRouter()}, srv: nil},
			args:    args{ctx: context.Background()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcfg := &newMuxConfig{
				serverCfg: tt.fields.serverCfg,
				router:    tt.fields.router,
				srv:       tt.fields.srv,
			}
			if err := mcfg.ShutDown(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("newMuxConfig.ShutDown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_gorillaRouter_pathPrefix(t *testing.T) {
	tt := &mux.Route{}
	tt.PathPrefix("/agent")
	type fields struct {
		router *mux.Router
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *mux.Route
	}{
		{
			name:   "Test path prefix",
			fields: fields{router: mux.NewRouter()},
			args:   args{path: "/agent"},
			want:   tt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grouter := &gorillaRouter{
				router: tt.fields.router,
			}
			//Route is internal struct and it has lot of field so not comparing. as a result checking nil
			if got := grouter.pathPrefix(tt.args.path); got == nil {
				t.Errorf("gorillaRouter.pathPrefix() = %+v, want %+v", *got, *tt.want)
			}
		})
	}
}

type sequencedHandler struct {
	next    http.Handler
	callSeq chan int
	num     int
}

func (h *sequencedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.callSeq <- h.num
	if h.next != nil {
		h.next.ServeHTTP(w, r)
	}
}

func (h *sequencedHandler) dummyMiddleware(next http.Handler) http.Handler {
	h.next = next
	return h
}

func TestGorillaRouter_Use_WithMiddleware(t *testing.T) {
	// given middleware funcs
	verificationOrder := make(chan int, 4)
	// create middleware with sequence numbers
	firstMiddleware := &sequencedHandler{callSeq: verificationOrder, num: 1}
	secondMiddleware := &sequencedHandler{callSeq: verificationOrder, num: 2}
	thirdMiddleware := &sequencedHandler{callSeq: verificationOrder, num: 3}
	// this is actual handler
	actualHandler := &sequencedHandler{callSeq: verificationOrder, num: 4}
	req, _ := http.NewRequest("GET", "/test", nil)
	rw := httptest.NewRecorder()
	router := &gorillaRouter{mux.NewRouter()}
	router.AddFunc("/test", actualHandler.ServeHTTP, http.MethodGet)
	router.Use(firstMiddleware.dummyMiddleware, secondMiddleware.dummyMiddleware, thirdMiddleware.dummyMiddleware)

	// when
	router.ServeHTTP(rw, req)
	close(verificationOrder)

	// expect middleware called in correct order 1,2,3
	assert.Equal(t, len(verificationOrder), 4)
	counter := 0
	for res := range verificationOrder {
		counter++
		assert.Equal(t, res, counter)
	}
}
