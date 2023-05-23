package webClient

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

type cbInfo struct {
	enabled      bool
	validCBError func(*http.Response) bool
}

// circuitBreaker hold circuit breaker command name and its enabled status
var circuitBreaker = make(map[string]cbInfo)

// CircuitBreakerConfig hold circuit breaker config for a host, service connects to
type CircuitBreakerConfig struct {

	// CircuitBreaker - Configuration for the Circuit breaker
	// Default config - circuitBreaker()
	CircuitBreaker *circuit.Config

	// BaseURL of the host for which circuit breaker needs to be configured
	BaseURL string

	// StateChangeCallback - callback to be used on circuit breaker state change
	// Default - logging the circuit breaker state change
	StateChangeCallback func(transaction string, commandName string, state string)

	// ValidCBError - Filter error reponse which counts towards circuit calculation
	// Default - statusCode >= 500
	ValidCBError func(*http.Response) bool
}

// RegisterCircuitBreaker - register circuit breaker for all hosts, service connects to.
// This function is not safe for concurrent use.
func RegisterCircuitBreaker(cb []CircuitBreakerConfig) error {

	circuit.Logger = logger.Get

	for _, cbCfg := range cb {
		if cbCfg.BaseURL != "" {

			u, err := url.Parse(cbCfg.BaseURL)
			if err != nil {
				return errors.Wrapf(err, "Failed to parse url: %s", cbCfg.BaseURL)
			}

			hostname := u.Host
			err = circuit.Register("", hostname, cbCfg.circuitBreaker(), cbCfg.StateChangeCallback)
			if err != nil {
				return errors.Wrapf(err, "Failed to register circuit breaker for url: %s", cbCfg.BaseURL)
			}
			circuitBreaker[hostname] = cbInfo{
				cbCfg.CircuitBreaker.Enabled,
				cbCfg.ValidCBError,
			}
		} else {
			return errors.New("Missing BaseURL in circuit breaker config")
		}
	}
	return nil
}

// circuitBreaker - set default config for circuit breaker
func (c *CircuitBreakerConfig) circuitBreaker() *circuit.Config {
	if c.CircuitBreaker == nil {
		c.CircuitBreaker = &circuit.Config{
			Enabled: true, TimeoutInSecond: 3, MaxConcurrentRequests: 15000,
			ErrorPercentThreshold: 25, RequestVolumeThreshold: 500, SleepWindowInSecond: 10,
		}
	}
	return c.CircuitBreaker
}
