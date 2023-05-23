package circuit

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/afex/hystrix-go/hystrix/callback"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
)

func TestCurrentState(t *testing.T) {
	type args struct {
		commandName string
	}
	tests := []struct {
		name  string
		args  args
		setup func()
		want  string
	}{
		{name: "1. State NA", want: "NA", args: args{commandName: "Invalid"}, setup: func() {}},
		{
			name: "2. State Closed", want: "Close", args: args{commandName: "Closed Command"},
			setup: func() { commandState["Closed Command"] = callback.Close },
		},
		{
			name: "3. State Open", want: "Open", args: args{commandName: "Open Command"},
			setup: func() { commandState["Open Command"] = callback.Open },
		},
		{
			name: "4. State AllowSingle", want: "Allow Single", args: args{commandName: "AllowSingle Command"},
			setup: func() { commandState["AllowSingle Command"] = callback.AllowSingle },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			if got := CurrentState(tt.args.commandName); got != tt.want {
				t.Errorf("CurrentState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	type args struct {
		transaction  string
		commandName  string
		config       *Config
		callbackFunc stateFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    string
	}{
		{name: "1. Invalid", wantErr: true, want: "NA", args: args{}},
		{
			name: "2. command exist", wantErr: false, want: "Close",
			args: args{commandName: "Test", config: &Config{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Register(tt.args.transaction, tt.args.commandName, tt.args.config, tt.args.callbackFunc); (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got := CurrentState(tt.args.commandName); got != tt.want {
				t.Errorf("Register() / CurrentState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stateChangeHandler(t *testing.T) {
	type args struct {
		transaction  string
		commandName  string
		state        callback.State
		callbackFunc stateFunc
	}
	tests := []struct {
		name  string
		args  args
		setup func()
	}{
		{
			name: "1. NA", args: args{commandName: "Invalid", state: callback.Open,
				callbackFunc: func(transaction string, commandName string, state string) {
					if state != "NA" && commandName != "Invalid" {
						t.Errorf("stateChangeHandler() = state: %v, want %v", state, "NA")
					}
				},
			},
			setup: func() {},
		},
		{
			name: "2. Close", args: args{commandName: "Closed Command", state: callback.Open,
				callbackFunc: func(transaction string, commandName string, state string) {
					if state != "Close" && commandName != "Closed Command" {
						t.Errorf("stateChangeHandler() = state: %v, want %v", state, "NA")
					}
				},
			},
			setup: func() { commandState["Closed Command"] = callback.Close },
		},
		{
			name: "3. Panic", args: args{commandName: "Closed Command", state: callback.Open,
				callbackFunc: func(transaction string, commandName string, state string) {
					panic("Error")
				},
			},
			setup: func() { commandState["Closed Command"] = callback.Close },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			stateChangeHandler(tt.args.transaction, tt.args.commandName, tt.args.state, tt.args.callbackFunc)
		})
	}
}

func TestDo(t *testing.T) {
	t.Run("Invalid Command - Circuit", func(t *testing.T) {
		got := Do("Invalid Command", true, func() error {
			return fmt.Errorf("Error Invalid Command")
		}, nil)

		if got == nil {
			t.Errorf("Do() = %v, want : Error Invalid Command", got)
		}
	})

	t.Run("Invalid Command - No Circuit", func(t *testing.T) {
		got := Do("Invalid Command", false, func() error {
			return fmt.Errorf("Error Invalid Command")
		}, nil)

		if got == nil {
			t.Errorf("Do() = %v, want : Error Invalid Command", got)
		}
	})

	t.Run("Test Command - Circuit Open", func(t *testing.T) {
		transaction := utils.GetTransactionID()
		testCommandName := fmt.Sprintf("Test Command %v", transaction)

		Register(transaction, testCommandName, New(), func(transaction string, commandName string, state string) {
			if commandName != testCommandName {
				t.Errorf("Do() = %v, want : %v", commandName, testCommandName)
			}
		})

		for index := 0; index < 30; index++ {
			got := Do(testCommandName, true, func() error {
				return fmt.Errorf("Error Test Command")
			}, nil)

			if index < 20 && (got == nil || !strings.Contains(got.Error(), "Error Test Command")) {
				t.Errorf("Do() %v = %v, want : Error Test Command", index, got)
			} else if index == 20 && (got == nil || !strings.Contains(got.Error(), hystrix.ErrCircuitOpen.Message)) {
				/* RequestVolumeThreshold expect 20 requests by default to come in before the circuit breaks.
				 * A race condition in Hystrix can occur where the metrics that track the volume
				 * have not finished updating by the time the 20th request checks if it is in a healthy state,
				 * causing a delay on the circuit opening and our subsequent tests to fail.
				 * Here we add a delay to let Hystrix finish updating so the circuit will break on the next request.
				 */
				time.Sleep(10 * time.Millisecond)
			} else if index > 20 && (got == nil || !strings.Contains(got.Error(), hystrix.ErrCircuitOpen.Message)) {
				t.Errorf("Do() %v = %v, want : circuit open", index, got)
			}
		}
	})
}

func TestGo(t *testing.T) {
	t.Run("Circuit enabled but closed", func(t *testing.T) {
		errs := Go("closed", true, func() error {
			return fmt.Errorf("Error Invalid Command")
		}, nil)

		assert.NotNil(t, errs)
	})

	t.Run("Circuit disabled", func(t *testing.T) {
		errs := Go("disabled", false, func() error {
			return fmt.Errorf("Error Invalid Command")
		}, nil)

		require.Len(t, errs, 1)
		err := <-errs
		assert.False(t, strings.Contains(err.Error(), ErrCircuitOpenMessage))
	})

	t.Run("Circuit Open", func(t *testing.T) {
		transaction := utils.GetTransactionID()
		err := Register(transaction, "open", New(), nil)
		require.NoError(t, err)

		for index := 0; index < 30; index++ {
			errs := Go("open", true, func() error {
				return fmt.Errorf("Error Test Command")
			}, nil)

			require.NotNil(t, errs)
			err = <-errs
			if index < 20 {
				assert.False(t, strings.Contains(err.Error(), ErrCircuitOpenMessage))
			} else if index == 20 && (err == nil || !strings.Contains(err.Error(), hystrix.ErrCircuitOpen.Message)) {
				/* RequestVolumeThreshold expect 20 requests by default to come in before the circuit breaks.
				 * A race condition in Hystrix can occur where the metrics that track the volume
				 * have not finished updating by the time the 20th request checks if it is in a healthy state,
				 * causing a delay on the circuit opening and our subsequent tests to fail.
				 * Here we add a delay to let Hystrix finish updating so the circuit will break on the next request.
				 */
				time.Sleep(10 * time.Millisecond)
			} else {
				assert.True(t, strings.Contains(err.Error(), ErrCircuitOpenMessage))
			}
		}
	})
}
