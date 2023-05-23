package db

import (
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

const (
	cbEnabled                = true
	cbTimeoutInSeconds       = 3
	cbMaxConcurrentRequests  = 15000
	cbErrorPercentThreshold  = 25
	cbRequestVolumeThreshold = 10
	cbSleepWindowInSecond    = 10
)

// Logger : Logger instance used for logging
// Defaults to Discard
var Logger = logger.DiscardLogger

// CircuitBreaker holds circuit breaker config for a database, the application connects to.
type CircuitBreaker struct {

	// Config - Configuration for the Circuit breaker
	// Default config - circuitBreaker()
	Config *circuit.Config

	// StateChangeCallback - callback to be used on circuit breaker state change
	// Default - logging the circuit breaker state change
	StateChangeCallback func(transaction string, commandName string, state string)
}

// Config is struct to define db configurations
type Config struct {
	//DbName - Db to be selected after connecting to server
	//Required
	DbName string

	//Server - ip address of db host server
	//Required
	Server string

	//UserID - UserId for db server
	//Required
	UserID string

	//Password - Password for db server
	//Required
	Password string

	//Driver - Name of db driver
	//Required
	Driver string

	//Map to hold additional db config
	AdditionalConfig map[string]string

	//CacheLimit - CacheLimit sets limit on number of prepared statements to be cached
	//Default CacheLimit: 100
	CacheLimit int

	// CircuitBreaker struct contains configuration values for database circuit breaker.
	CircuitBreaker CircuitBreaker
}

// CircuitBreaker - set default config for circuit breaker.
func (c *CircuitBreaker) circuitBreaker() *circuit.Config {
	if c.Config == nil {
		c.Config = &circuit.Config{
			Enabled: cbEnabled, TimeoutInSecond: cbTimeoutInSeconds, MaxConcurrentRequests: cbMaxConcurrentRequests,
			ErrorPercentThreshold: cbErrorPercentThreshold, RequestVolumeThreshold: cbRequestVolumeThreshold, SleepWindowInSecond: cbSleepWindowInSecond,
		}

		return c.Config
	}

	return c.Config
}
