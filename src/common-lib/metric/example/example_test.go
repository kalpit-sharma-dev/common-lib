package main

import (
	"errors"
	"testing"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/metric"
)

func Test_publish(t *testing.T) {
	type args struct {
		namespace  string
		address    string
		portNumber string
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "1 metric.Publish Error",
			args: args{namespace: "", address: "", portNumber: ""},
			setup: func() {
				metric.Publish = func(cfg *metric.Config, collector ...metric.Collector) error { return errors.New("Error") }
				metric.PeriodicPublish = func(duration time.Duration, cfg *metric.Config, callback func() []metric.Collector, handler func(err error)) {
				}
			},
			wantErr: true,
		},
		{
			name: "1 metric.Publish Success",
			args: args{namespace: "", address: "", portNumber: ""},
			setup: func() {
				metric.Publish = func(cfg *metric.Config, collector ...metric.Collector) error { return nil }
				metric.PeriodicPublish = func(duration time.Duration, cfg *metric.Config, callback func() []metric.Collector, handler func(err error)) {
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			if err := publish(tt.args.namespace, tt.args.address, tt.args.portNumber); (err != nil) != tt.wantErr {
				t.Errorf("publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
