package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/util"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/tracing"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/webClient"
)

func main() {

	// setup xray configuration (use defaults)
	tracingConfig := tracing.NewConfig()
	tracingConfig.ServiceVersion = "1.0.0"
	tracingConfig.ServiceName = util.Hostname(util.ProcessName()) + "_webserver1"

	// printing to know what to look for in xray console
	fmt.Println(tracingConfig.ServiceName)

	// get the router and register route
	port := ":8080"
	server := web.Create(&web.ServerConfig{ListenURL: port, TracingConfig: tracingConfig})
	r := server.GetRouter()
	r.AddFunc("/webserver1", func(w http.ResponseWriter, r *http.Request) {
		// make downstream api request with xray context
		makeAPIRequestWithContext(r.Context(), tracingConfig)

		// respond to client
		w.Header().Add("Content-Type", "application/json")
		e := json.NewEncoder(w)
		err := e.Encode(map[string]string{"hello": "webserver1 made request to webserver2"})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("finished request on webserver1")
	}, http.MethodGet)

	// start the server
	ctx1, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fmt.Printf("webserver1 listening on port %s\n", port)
	server.Start(ctx1)
}

func makeAPIRequestWithContext(ctx context.Context, tracingCfg *tracing.Config) {
	// setup client config
	config := webClient.ClientConfig{}
	config.TracingConfig = tracingCfg

	// get client
	factory := webClient.ClientFactoryImpl{}
	client := factory.GetClientServiceByType(webClient.BasicClient, config)

	// prepare request
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8090/webserver2", nil)

	// make request
	res, err := client.Do(req)

	// parse result
	if err != nil {
		fmt.Println(err)
	} else {
		defer res.Body.Close()
		resData, _ := ioutil.ReadAll(res.Body)
		resStr := string(resData)
		fmt.Printf("webserver1 response: %s \n", resStr)
	}
}
