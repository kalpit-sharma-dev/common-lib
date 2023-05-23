package circuit

import "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"

// Logger : Logger instance used for logging
// Defaults to Discard
var Logger = logger.DiscardLogger

// Config - All the Circuit breaker related configurations
type Config struct {
	// Enabled - Circuit breaker is enabled or not
	// Default - Enabled is true
	Enabled bool

	// TimeoutInSecond - How long to wait for command to complete, in Seconds
	// DefaultTimeout = 1
	TimeoutInSecond int

	// How many commands of the same type can run at the same time
	// MaxConcurrentRequests - DefaultMaxConcurrent = 10
	MaxConcurrentRequests int

	// ErrorPercentThreshold - Causes circuits to open once the rolling measure of errors exceeds this percent of requests
	// DefaultErrorPercentThreshold = 50
	ErrorPercentThreshold int

	// RequestVolumeThreshold - RequestVolumeThreshold - The minimum number of requests in the rolling window (10 Sec)
	// after which the error percent will be calculated.
	// DefaultVolumeThreshold = 20
	RequestVolumeThreshold int

	// SleepWindowInSecond - How long, in Seconds, to wait after a circuit opens before testing for recovery
	// DefaultSleepWindow = 5
	SleepWindowInSecond int
}

// New - Creates a defaut config object and returns a value
func New() *Config {
	return &Config{
		Enabled: true, TimeoutInSecond: 1, MaxConcurrentRequests: 10,
		ErrorPercentThreshold: 50, RequestVolumeThreshold: 20, SleepWindowInSecond: 5,
	}
}
