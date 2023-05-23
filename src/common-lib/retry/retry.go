package retry

import (
	"time"

	retry "github.com/avast/retry-go"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

// DelayType defines the type fo delay to be used
type delayType int

const (
	//FixedDelay use constant delay strategy between consecutive retries
	FixedDelay delayType = iota
	//ExponentialDelay uses exponential back-off delay strategy between consecutive retries
	//Please note that exponential delay will keep retrying with exponential increase in duration between retries
	//this means that nth retry (excluding first try) would be executing (2^(n-1) * delay duration) after the (n-1)th retry
	ExponentialDelay
)

// Task is the main struct for implementing the retry mechanism in your code
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

// Logger to be used for generic logging purposes
var Logger = logger.DiscardLogger

// NewTask initializes the retry config with the default values.
// This may be typically used to create initial config and override specific values using the setters provided
func NewTask() Task {
	rc := Task{
		Attempts:              3,
		BaseDelay:             10 * time.Second,
		DelayType:             FixedDelay,
		ExecuteFunctionOnFail: nil,
		CheckIfRetry:          retry.DefaultRetryIf,
	}
	return rc
}

func getDelayMethod(delayType delayType) retry.DelayTypeFunc {
	switch delayType {
	case FixedDelay:
		return retry.FixedDelay
	case ExponentialDelay:
		return retry.BackOffDelay
	default:
		return retry.FixedDelay
	}
}

// Do the main retry function
func (rc Task) Do(task func() error, transactionID string) error {
	if rc.ExecuteFunctionOnFail == nil {
		var callback = func(attemptNum uint, err error) {
			Logger().Error(transactionID, "platform-common-lib.retry.retry.1", "Error for attempt=%v is err=%v", attemptNum, err)
		}

		rc.ExecuteFunctionOnFail = callback
	}

	err := retry.Do(task,
		retry.Attempts(rc.Attempts),
		retry.Delay(rc.BaseDelay),
		retry.OnRetry(rc.ExecuteFunctionOnFail),
		retry.RetryIf(rc.CheckIfRetry),
		retry.DelayType(getDelayMethod(rc.DelayType)))
	return err
}
