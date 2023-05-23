package with

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestContext(t *testing.T) {
	minuteCtx, cm := context.WithTimeout(context.Background(), time.Minute)
	nenoCtx, cn := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cm()
	defer cn()

	type args struct {
		ctx         context.Context
		name        string
		transaction string
		fn          func() error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "1 Function Error", wantErr: true, args: args{ctx: minuteCtx, name: "1", transaction: "1", fn: func() error { return errors.New("Error") }}},
		{name: "2 Function Panic", wantErr: true, args: args{ctx: minuteCtx, name: "1", transaction: "1", fn: func() error { panic("Error") }}},
		{name: "3 Timeout", wantErr: true, args: args{ctx: nenoCtx, name: "1", transaction: "1", fn: func() error { return errors.New("Error") }}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Context(tt.args.ctx, tt.args.name, tt.args.transaction, tt.args.fn); (err != nil) != tt.wantErr {
				t.Errorf("Context() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
