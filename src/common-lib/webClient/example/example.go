package main

import (
	"fmt"
	"net/http"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/webClient"
)

func main() {
	// Set circuit breaker config
	h := webClient.CircuitBreakerConfig{
		CircuitBreaker: circuit.New(),
		BaseURL:        "http://localhost:9090/hello",
	}

	s := []webClient.CircuitBreakerConfig{h}

	// Register circuit breaker
	err := webClient.RegisterCircuitBreaker(s)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Getting health check implementation for each host
	_, err = webClient.Health(s)
	if err != nil {
		fmt.Println("Error getting health status: " + err.Error())
	}
	clientFact := webClient.ClientFactoryImpl{}
	clientService := clientFact.GetClientServiceByType(webClient.TLSClient, webClient.ClientConfig{})

	for index := 0; index < 100; index++ {
		req, _ := http.NewRequest(http.MethodGet, "http://localhost:9090", nil)
		clientService.Do(req)
	}
	fmt.Printf("Circuit State: %v", circuit.CurrentState("localhost:9090"))
}
