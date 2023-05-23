package with

import (
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

func TestRecover(t *testing.T) {
	Log, _ = logger.Create(logger.Config{Name: "ExecuteWith", LogLevel: logger.OFF, Destination: logger.DISCARD})
	type args struct {
		name        string
		transaction string
		fn          func()
		handler     func(transaction string, err error)
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "1-excutePanic", args: args{name: "1", transaction: "1", fn: func() { panic("Error") }}},
		{name: "2-excutePanicHandlerPanic", args: args{name: "2", transaction: "1", fn: func() { panic("Error") },
			handler: func(transaction string, err error) { panic("Error") }}},
		{name: "3-excutePanicHandlerSuccess", args: args{name: "3", transaction: "1", fn: func() { panic("Error") },
			handler: func(transaction string, err error) {}}},
		{name: "4-excuteSuccessNoHandler", args: args{name: "4", transaction: "1", fn: func() {}}},
		{name: "5-excuteSuccessHandler", args: args{name: "5", transaction: "1", fn: func() {},
			handler: func(transaction string, err error) { t.Errorf("Error") }}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Recover(tt.args.name, tt.args.transaction, tt.args.fn, tt.args.handler)
		})
	}
}
