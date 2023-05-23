package tracing

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestTracing(t *testing.T) {
	tests := []struct {
		name string
		arg  *Config
	}{
		{name: "default", arg: &Config{
			Enabled: false,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configure(tt.arg)
			handler := http.HandlerFunc(nil)
			WrapHandlerWithTracing(tt.arg, handler)
			client := &http.Client{}
			HTTPClient(tt.arg, client)
		})
	}
}

func TestCapture(t *testing.T) {
	old := traceEnable
	traceEnable = func() bool {
		return false
	}
	defer func() {
		traceEnable = old
	}()

	errMsg := "tracing not enabled"
	err := Capture(context.Background(), "testSegment", func(ctx context.Context) error {
		return errors.New(errMsg)
	})

	if err.Error() != errMsg {
		t.Error("Tracing is disabled but called function not getting called")
	}
}
