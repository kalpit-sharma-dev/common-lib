<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Circuit Breaker

This is a Standard circuit breaker implementation used by all the Go projects in the google.

### Third-Party Libraties

- [hystrix](https://github.com/afex/hystrix-go/hystrix)
- **License** [MIT License](https://github.com/afex/hystrix-go/blob/master/LICENSE)
- **Description**
  - Hystrix is a latency and fault tolerance library designed to isolate points of access to remote systems, services and 3rd party libraries, stop cascading failure and enable resilience in complex distributed systems where failure is inevitable.
  - We have added an additional functionality in the Hystrix library by forking this.
- **Glide Dependencies**

```yaml
package: github.com/afex/hystrix-go
version: 4f7f0a216ae56fb0ef5f800521762de684b5598f
repo: https://github.com/googleLLC/hystrix-go
```

### [Example](example/example.go)

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
```

**Registration**

```go
/*
New - Creates a new circuit breaker configuration (i.e. object of 'Config') using the default values and returns pointer to this object. The default values are mentioned towards the end of this page.
This is optional quick way of creating circuit breaker configuration if default values are suitable. 
In case specific configuration is needed, please create object of 'Config' with required values
*/
config := circuit.New()

/* Registor a command for the circuit breaker. 
This function takes following four parameters
1. transaction - string : the unique transaction id used for logging the circuit breaker registration logs
2. commandName - string : the unique name for this circuit breaker. Multiple circuit breakers can be created with different names and can be uniquely identified using their name. This name shall be used while calling the 'Do' function
3. config - pointer to 'Config' object : this is pointer to the 'Config' object containing all configuration for this circuit breaker
4. callback - function : this is callback function for listnening to circuit breaker state change events. Whenever the state of circuit changes (e.g. from Close to HalfOpen)
*/
circuit.Register(transaction, commandName, config, func(transaction string, commandName string, state string))
```

**Calling Function in a synchronous manner**

```go
/* Do - To be called for synchronous Circuit breaker execution
Parameters are 
1. commandName - string:  the unique name for this circuit breaker used while registering the circuit breaker
2. circuitEnabled - bool : flag to indicate if circuit breaker should be used or not. False indicates the circuit breaker should not be used for the call
3. execute - function : the execute function which is executed by circuit breaker. This is typically a function closure as the function does not accept parameters and returns only error during execution i.e. no response
4. fallback - function : the fallback function is called whenever execute function execution results in error. If this fallback does not return error, then the call is assumed to be completed without error
*/
err := circuit.Do(commandName, true, func() error {
	return fmt.Errorf("Error")
}, nil)

if err != nil {
	logger.Get().Error(transaction, "Error", "%v", err)
}

// CurrentState - To return Current state of a command
circuit.CurrentState(commandName string) string
```

**Calling Function in a concurrent manner**

```go
/* Go - To be called for Circuit breaker execution with concurrent tracking of the health of previous calls to the function
Parameters are 
1. commandName - string:  the unique name for this circuit breaker used while registering the circuit breaker
2. circuitEnabled - bool : flag to indicate if circuit breaker should be used or not. False indicates the circuit breaker should not be used for the call
3. execute - function : the execute function which is executed by circuit breaker. This is typically a function closure as the function does not accept parameters and returns only error during execution i.e. no response
4. fallback - function : the fallback function is called whenever execute function execution results in error. If this fallback does not return error, then the call is assumed to be completed without error
*/
errChan := circuit.Go(commandName, true, func() error {
	return fmt.Errorf("Error")
}, nil)


if errChan != nil {
	logger.Get().Error(transaction, "Error", "%v", <-errChan)
}

// CurrentState - To return Current state of a command
circuit.CurrentState(commandName string) string
```

**Constants and Errors**

```go
// Open - state to indicate that Circuit state is Open
circuit.Open = "Open"

// Close - state to indicate that Circuit state is Close
circuit.Close = "Close"

// HalfOpen - state to indicate that Circuit state is HalfOpen or trying to Open Circuit
circuit.HalfOpen = "Half-Open"

// NA - state to indicate that Circuit state is Not Available
circuit.NA = "NA"

// ErrNilCommandName - if user does not provide command name while registration this will be returned.
circuit.ErrNilCommandName = errors.New("NilCommandName; Prvoide a unique name for registration")

// Logger : Logger instance used for logging
// Defaults to Discard
circuit.Logger = logger.DiscardLogger
```

**Configuration**

```go
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

	// RequestVolumeThreshold - The minimum number of requests in the rolling window (10 Sec)
	// after which the error percent will be calculated.
	// DefaultVolumeThreshold = 20
	RequestVolumeThreshold int

	// SleepWindowInSecond - How long, in Seconds, to wait after a circuit opens before testing for recovery
	// DefaultSleepWindow = 5
	SleepWindowInSecond int
}
```

## Note:
- We need to define all these values inside a rolling window
  - Current rolling window is 10 Second and non configurable


### Contribution

Any changes in this package should be communicated to Juno Team.
