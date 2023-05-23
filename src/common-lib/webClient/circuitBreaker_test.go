package webClient

import (
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

func TestRegisterCircuitBreaker(t *testing.T) {
	t.Run("1. Error parsing url", func(t *testing.T) {

		cbCfg := CircuitBreakerConfig{
			CircuitBreaker: circuit.New(),
			BaseURL:        "%gh&%ij",
		}

		err := RegisterCircuitBreaker([]CircuitBreakerConfig{cbCfg})
		if err == nil {
			t.Error("Expecting error but found nil")
		}
	})

	t.Run("2. Missing Base URL", func(t *testing.T) {

		cbCfg := CircuitBreakerConfig{
			CircuitBreaker: circuit.New(),
		}

		err := RegisterCircuitBreaker([]CircuitBreakerConfig{cbCfg})
		if err == nil {
			t.Error("Expecting error but found nil")
		}
	})

	t.Run("3. Success", func(t *testing.T) {
		hostName := "www.google.com"
		cbCfg := CircuitBreakerConfig{
			BaseURL: "http://www.google.com",
		}

		err := RegisterCircuitBreaker([]CircuitBreakerConfig{cbCfg})
		if err != nil {
			t.Errorf("Expecting nil but found err: %v", err)
		}

		if cbInfo, ok := circuitBreaker[hostName]; !ok || !cbInfo.enabled {
			t.Errorf("Expecting circuit breaker to be enabled but found disabled")
		}

		if state := circuit.CurrentState(hostName); state != circuit.Close {
			t.Errorf("Expecting circuit state to be closed but found: %v", state)
		}
	})
}
