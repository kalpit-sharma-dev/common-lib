package udp

import (
	"errors"
	"testing"
	"time"
)

func TestSend(t *testing.T) {
	type args struct {
		conf    *Config
		message []byte
		handler ResponseHandler
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "1 Wrong Address", wantErr: true,
			args: args{conf: &Config{TimeoutInSeconds: 1, Address: "a.b.c.d"}, message: []byte("Test")},
		},
		{
			name: "2 Handler Error", wantErr: true,
			args: args{conf: &Config{TimeoutInSeconds: 1, Address: "192.168.0.1"}, message: []byte("Test"),
				handler: func(response Response) error { return errors.New("Error") }},
		},
		{
			name: "3 Handler Panic", wantErr: true,
			args: args{conf: &Config{TimeoutInSeconds: 10}, message: []byte("Test"),
				handler: func(response Response) error { panic(errors.New("Error")) }},
		},
		{
			name: "4 Handler Timeout", wantErr: true,
			args: args{conf: &Config{TimeoutInSeconds: 1, Address: "192.168.0.1"}, message: []byte("Test"),
				handler: func(response Response) error { time.Sleep(2 * time.Second); return nil }},
		},
		{
			name: "5 Succes", wantErr: false,
			args: args{conf: &Config{TimeoutInSeconds: 2, Address: "192.168.0.1"}, message: []byte("Test")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Send(tt.args.conf, tt.args.message, tt.args.handler); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
