package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/tracing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/util"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
)

func main() {

	// setup xray configuration (use defaults)
	tracingConfig := tracing.NewConfig()
	tracingConfig.ServiceVersion = "1.0.0"
	tracingConfig.ServiceName = util.Hostname(util.ProcessName()) + "_webserver2"

	// printing to know what to look for in xray console
	fmt.Println(tracingConfig.ServiceName)

	// get the router and register route
	port := ":8090"
	server := web.Create(&web.ServerConfig{ListenURL: port, TracingConfig: tracingConfig})
	r := server.GetRouter()
	r.AddFunc("/webserver2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		e := json.NewEncoder(w)
		err := e.Encode(map[string]string{"hello": "webserver2"})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("finished request on webserver2")
	}, http.MethodGet)

	// start the server
	ctx1, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fmt.Printf("webserver2 listening on port %s\n", port)
	server.Start(ctx1)
}
