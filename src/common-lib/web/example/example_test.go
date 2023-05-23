package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	cweb "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/mock"
)

func Test_main(t *testing.T) {
	ctrl := gomock.NewController(t)

	originalcreateServer := createServer
	defer func() {
		ctrl.Finish()
		createServer = originalcreateServer
	}()
	mockServer := mock.NewMockHTTPServer(ctrl)
	mockRouter := mock.NewMockRouter(ctrl)

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "To test main functions",
			setup: func() {
				createServer = func(cfg *cweb.ServerConfig) cweb.HTTPServer {
					return mockServer
				}
				mockServer.EXPECT().GetRouter().Return(mockRouter).AnyTimes()
				mockRouter.EXPECT().AddFunc(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				mockRouter.EXPECT().AddHandle(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				mockRouter.EXPECT().Use(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				mockServer.EXPECT().Start(gomock.Any()).AnyTimes()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			setup()
		})
	}
}
func Router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/apath-to-api", HandlerFuncWhenwithoutmiddlware).Methods("GET")
	router.HandleFunc("/path-to-api", AuthMiddleware(HandlerFuncWhenMiddleware)).Methods("GET")
	return router
}

func TestEndpoint(t *testing.T) {
	request, _ := http.NewRequest("GET", "/apath-to-api", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")

	request, _ = http.NewRequest("GET", "/path-to-api", nil)
	response = httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
}

func TestMiddleware(t *testing.T) {
	// given
	request, _ := http.NewRequest("GET", "/path-to-api", nil)
	response := httptest.NewRecorder()
	r := Router()
	tm := &testMiddleware{}
	r.Use(tm.dummyMiddleware)
	// when
	r.ServeHTTP(response, request)
	// expect middleware has been invoked
	assert.Equal(t, response.Body.String(), "test")
}
