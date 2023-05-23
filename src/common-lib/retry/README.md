<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Generic Retry Framework

This implementation allows the end user to retry a function or a block of code.
The user can use the implementation out of the box or tailor it as per his needs. 

### Third-Party Libraries

- [avast/retry-go](https://github.com/avast/retry-go)
- **License** [MIT License](https://github.com/avast/retry-go/blob/master/LICENSE)
- **Description**
  - Simple library for retry mechanism
- **Glide Dependencies**

```yaml
package: github.com/avast/retry-go
version: a8f6dc7e8f46a5d11c01d9da1e887bb5a51fd107
```

### [Example]

**Import Statement**

```go
import	retry "github.com/ContinuumLLC/platform-common-lib/src/retry"
```

**Registration**

```go
//NewTask initializes the retry config with the default values.
//This may be typically used to create initial config and override specific values
func NewTask() Task


//In order for client to execute any additional functionality that he wants as per his needs
//they should set the value of task variable ExecuteFunctionOnFail
rc.ExecuteFunctionOnFail = callback
```

**Calling Function**

```go
//Execute the main retry function and retry based on Provided Configuration.
//Typically the task is a 'closure' as it does not accept any parameter and 
//returns only error.
//This method will block until all retries are executed or next retry is denied (using CheckIfRetry) 
//It returns error with string details containing list of all errors occurred in all retries 
//iff no retry is successful
//When any retry succeeds, the methods returns nil 
//transactionid passed here will be used for logging attempt number and error by default if no custom
//callback function is provided as described above
func (rc Task) Do(task func() error, transactionID string) error {
	if rc.ExecuteFunctionOnFail == nil {
		var callback = func(attemptNum uint, err error) {
			Logger().Error(transactionID, "platform-common-lib.retry.retry.1", "Error for attempt=%v is err=%v", attemptNum, err)
		}

		rc.ExecuteFunctionOnFail = callback
	}

	// main logic starts below
} 
```

**Enums**

```go
//delayType defines the type fo delay to be used
type delayType int

const (
	//FixedDelay use constant delay strategy between consecutive retries
	//This is also the default value used in NewTask()
	FixedDelay delayType = iota
	
	//ExponentialDelay uses exponential back-off delay strategy between consecutive retries
	//Please note that exponential delay will keep retrying with exponential increase in duration between retries
	//this means that nth retry (excluding first try) would be executing (2^(n-1) * delay duration) after the (n-1)th retry
	ExponentialDelay
)
```

**Configuration**

```go
// Task - All the retry related configurations
type Task struct {
	//Attempts the maximum number of attempts to be retries for provided task
	Attempts uint

	//BaseDelay the base delay duration to be used between retry attempts of provided task.
	//The actual delay depends on delayType.
	//See documentation of delayType enumeration
	BaseDelay time.Duration

	//ExecuteFunctionOnFail the callback function called when any try fails
	ExecuteFunctionOnFail func(attemptNum uint, err error)

	//CheckIfRetry the callback function called to check if the next retry should be performed
	CheckIfRetry func(err error) bool

	//DelayType the delay type to be used from enumeration of delay types
	DelayType delayType
}

//Logger to be used for generic logging purposes
//This variable can be pointed to your local logger init and can be used
//which will be log the attempt number and the error for that attempt 
var Logger = logger.DiscardLogger
```

### Contribution

All code changes in this package should be communicated to Juno Team.
