package producer

import (
	"context"
	"testing"
	"time"
)

func Test_NewSyncProducer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "default",
		},
	}
	for _, tt := range tests {
		c := NewConfig()
		t.Run(tt.name, func(t *testing.T) {
			c.Address = []string{"localhost"}
			p, err := NewSyncProducer(c)
			if err != nil {
				t.Errorf("NewSyncProducer() errored")
			}
			p.Close()
		})
	}
}

func Test_SyncProducerHealth(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{
			name: "default",
			want: true, // it starts off immediately as true until lib connects to brokers and fails
		},
	}
	for _, tt := range tests {
		c := NewConfig()
		t.Run(tt.name, func(t *testing.T) {
			c.Address = []string{"localhost"}
			p, err := NewSyncProducer(c)
			if err != nil {
				t.Errorf("NewSyncProducer() errored")
			}
			h, err := p.Health()
			if h == nil {
				t.Errorf("Health() returned nil")
			}
			if h.ConnectionState != tt.want {
				t.Errorf("Health().ConnectionState expected %t, got %t", tt.want, h.ConnectionState)
			}
		})
	}
}

func Test_SyncProducerconnStateChange(t *testing.T) {
	tests := []struct {
		name string
		arg  bool
		want bool
	}{
		{
			name: "true",
			arg:  true,
			want: true,
		},
		{
			name: "false",
			arg:  false,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			config.Address = []string{"localhost"}
			p, err := NewSyncProducer(config)
			if err != nil {
				t.Errorf("NewSyncProducer() errored")
			}
			ap := p.(*syncProducer)
			ap.connStateChange(tt.arg)
			h, err := p.Health()
			if h == nil {
				t.Errorf("Health() returned nil")
			}
			if h.ConnectionState != tt.want {
				t.Errorf("Health().ConnectionState expected %t, got %t", tt.want, h.ConnectionState)
			}
		})
	}
}

func Test_ProduceWithReportErr(t *testing.T) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	tests := []struct {
		name string
		arg  []*Message
		ctx  context.Context
		want error
	}{
		{
			name: "ErrPublishMessageNotAvailable",
			arg:  []*Message{},
			ctx:  context.Background(),
			want: ErrPublishMessageNotAvailable,
		},
		{
			name: "ErrPublishSendMessageTimeout",
			arg: []*Message{
				{Topic: "test", Value: []byte("value")},
			},
			ctx:  timeoutCtx,
			want: ErrPublishSendMessageTimeout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			c.Address = []string{"localhost"}
			p, err := NewSyncProducer(c)
			if err != nil {
				t.Errorf("NewSyncProducer() errored")
			}
			_, err = p.ProduceWithReport(tt.ctx, "", tt.arg...)
			if err.Error() != tt.want.Error() {
				t.Errorf("ProduceWithReport() expected %v, got %v", tt.want.Error(), err.Error())
			}
			p.Close()
		})
	}
}
