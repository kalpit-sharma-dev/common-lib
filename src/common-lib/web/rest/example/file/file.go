package main

import (
	"encoding/json"
	"net/http"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
)

func main() {
	createMuxServer()
}

func createMuxServer() {
	server := web.ServerFactoryImpl{}.GetServer(&web.ServerConfig{
		URLPathPrefix:        "/v1",
		ListenURL:            "localhost:8888",
		ReadTimeoutMinute:    1,
		WriteTimeoutMinute:   1,
		IdleTimeoutMinute:    1,
		MaxHandlers:          2,
		MaxConcurrentStreams: 3,
		StaticFileDirectory:  ".",
	})
	server.SetupRoutes([]*web.RouteConfig{
		{
			URLPathSuffix: "/test",
			Res:           testResource{},
		},
	})
	server.ListenAndServe()
}

type testResource struct {
	web.Post405
	web.Put405
	web.Delete405
	web.Others405
}

func (res testResource) Get(rc web.RequestContext) {
	w := rc.GetResponse()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-type", "application/json")
	err := json.NewEncoder(w).Encode("Test Resource excecution")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
