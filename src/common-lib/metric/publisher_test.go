package metric

import (
	"errors"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/udp"
)

func TestPublish(t *testing.T) {
	type args struct {
		cfg       *Config
		collector []Collector
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "1 Publish - No Collector - Communication Error",
			args: args{cfg: &Config{Communication: &udp.Config{}}, collector: nil},
			setup: func() {
				udp.Send = func(conf *udp.Config, message []byte, handler udp.ResponseHandler) error { return errors.New("Error") }
			},
			wantErr: true,
		},
		{
			name: "Publish - Collector - Communication Error",
			args: args{cfg: &Config{Communication: &udp.Config{}}, collector: []Collector{&Counter{}}},
			setup: func() {
				udp.Send = func(conf *udp.Config, message []byte, handler udp.ResponseHandler) error { return errors.New("Error") }
			},
			wantErr: true,
		},
		{
			name: "1 Publish - No Collector - Communication Success",
			args: args{cfg: &Config{Communication: &udp.Config{}}, collector: nil},
			setup: func() {
				udp.Send = func(conf *udp.Config, message []byte, handler udp.ResponseHandler) error { return nil }
			},
			wantErr: false,
		},
		{
			name: "Publish - Collector - Communication Success",
			args: args{cfg: &Config{Communication: &udp.Config{}}, collector: []Collector{&Counter{}}},
			setup: func() {
				udp.Send = func(conf *udp.Config, message []byte, handler udp.ResponseHandler) error { return nil }
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			if err := Publish(tt.args.cfg, tt.args.collector...); (err != nil) != tt.wantErr {
				t.Errorf("Publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_publish(t *testing.T) {
	type args struct {
		message *Message
		cfg     *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := publish(tt.args.message, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
