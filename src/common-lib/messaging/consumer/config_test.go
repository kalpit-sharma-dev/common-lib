package consumer

import (
	"reflect"
	"testing"
	"time"
)

func TestNewConfigDefaults(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		{name: "default", want: &Config{
			SubscriberPerCore:    20,
			CommitMode:           OnPull,
			ConsumerMode:         PullUnOrdered,
			OffsetsInitial:       OffsetNewest,
			Timeout:              time.Minute,
			ErrorHandlingTimeout: time.Minute,
			RetryCount:           10,
			RetryDelay:           30 * time.Second,
			Partitions:           500,
			MaxQueueSize:         100,
			EmptyQueueWaitTime:   500 * time.Millisecond,
			CommitIntervalMs:     5000,
			EnableKafkaLogs:      false,
			LogLevel:             3,
			LoggedComponents:     "all",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
