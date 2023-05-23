package circuit

import (
	"errors"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/afex/hystrix-go/hystrix/callback"
)

const (
	// Open - state to indicate that Circuit state is Open
	Open = "Open"

	// Close - state to indicate that Circuit state is Close
	Close = "Close"

	// HalfOpen - state to indicate that Circuit state is HalfOpen or trying to Open Circuit
	HalfOpen = "Half-Open"

	// NA - state to indicate that Circuit state is Not Available
	NA = "NA"
)

var (
	// ErrNilCommandName - if user does not provide command name while registration this will be returned.
	ErrNilCommandName = errors.New("NilCommandName; Prvoide a unique name for registration")

	// ErrCircuitOpenMessage - circuit open error message
	ErrCircuitOpenMessage = hystrix.ErrCircuitOpen.Message

	// ErrMaxConcurrencyMessage - circuit max concurrency error message
	ErrMaxConcurrencyMessage = hystrix.ErrMaxConcurrency.Message

	// ErrTimeoutMessage - circuit timeout error message
	ErrTimeoutMessage = hystrix.ErrTimeout.Message

	commandState = make(map[string]callback.State)

	stateChangeMutex sync.RWMutex
)

type stateFunc func(transaction string, commandName string, state string)

// Do - To be called for Circuit breaker execution
var Do = func(commandName string, circuitEnabled bool, execute func() error, fallback func(error) error) error {
	if circuitEnabled {
		return hystrix.Do(commandName, execute, fallback)
	}
	return execute()
}

// Go - To be called for Circuit breaker execution to allow concurrent function execution
var Go = func(commandName string, circuitEnabled bool, execute func() error, fallback func(error) error) chan error {
	if circuitEnabled {
		return hystrix.Go(commandName, execute, fallback)
	}
	errors := make(chan error, 1)
	err := execute()
	if err != nil {
		errors <- err
	}
	return errors
}

// CurrentState - To return Current state of a command
func CurrentState(commandName string) string {
	stateChangeMutex.RLock()
	state, ok := commandState[commandName]
	stateChangeMutex.RUnlock()
	if !ok {
		return NA
	}
	return string(state)
}

// Register - To register callback function for a given command
func Register(transaction string, commandName string, config *Config, callbackFunc stateFunc) error {
	if strings.TrimSpace(commandName) == "" {
		return ErrNilCommandName
	}

	hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
		//Timeout and SleepWindow is in Milliseconds for hystrix -
		// DefaultTimeout = 1000, DefaultSleepWindow = 5000
		Timeout:                config.TimeoutInSecond * 1000,
		MaxConcurrentRequests:  config.MaxConcurrentRequests,
		ErrorPercentThreshold:  config.ErrorPercentThreshold,
		RequestVolumeThreshold: config.RequestVolumeThreshold,
		SleepWindow:            config.SleepWindowInSecond * 1000,
	})

	commandState[commandName] = callback.Close
	callback.Register(commandName, func(commandName string, state callback.State) {
		stateChangeHandler(transaction, commandName, state, callbackFunc)
	})
	Logger().Info(transaction, "Circuit breaker initialized => %s : %+v", commandName, hystrix.GetCircuitSettings()[commandName])
	return nil
}

func stateChangeHandler(transaction string, commandName string, state callback.State, callbackFunc stateFunc) {
	defer func() {
		if r := recover(); r != nil {
			Logger().Error(transaction,
				"StateChangeHandlerRecovered", "%s : StateChangeHandler : Recovered from %s Trace is : %s",
				commandName, r, debug.Stack())
		}
	}()

	currentState := CurrentState(commandName)
	stateChangeMutex.Lock()
	commandState[commandName] = state
	stateChangeMutex.Unlock()
	Logger().Info(transaction, "%s : Switching State from : %s ==> %s", commandName, currentState, state)

	if callbackFunc != nil {
		callbackFunc(transaction, commandName, string(state))
	}
}
