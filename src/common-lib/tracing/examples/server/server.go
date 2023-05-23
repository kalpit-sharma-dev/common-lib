package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/tracing"
)

func main() {
	tracingConfig := tracing.NewConfig()
	tracingConfig.Address = "localhost:2000"
	tracingConfig.ServiceVersion = "1.0.0"
	tracingConfig.ServiceName = "testApp"

	// configure deamon address and service version
	tracing.Configure(tracingConfig)

	// set routes
	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		hw := struct {
			Hello string `json:"hello"`
		}{Hello: "world!"}
		json.NewEncoder(w).Encode(hw)
	})

	// wrap http.Handler (router) with xray tracing middleware
	port := 8080
	handler, _ := tracing.WrapHandlerWithTracing(tracingConfig, router)
	log.Printf("Listening on :%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
}
