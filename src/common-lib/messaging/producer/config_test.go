package producer

import (
	"reflect"
	"testing"
)

func TestNewConfigDefaults(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		{
			name: "default",
			want: &Config{
				TimeoutInSecond:           5,
				CompressionType:           CompressionNone,
				MaxMessageBytes:           1000000,
				ReconnectBackoffMs:        100,
				ReconnectBackoffMaxMs:     10000,
				MessageTimeoutMs:          300000,
				MessageSendMaxRetries:     2,
				RetryBackoffMs:            100,
				QueueBufferingMaxMessages: 100000,
				QueueBufferingMaxKbytes:   1048576,
				QueueBufferingMaxMs:       5,
				EnableIdempotence:         false,
				ProduceChannelSize:        1000000,
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

func Test_CompressionConsts(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "CompressionNone",
			arg:  CompressionNone,
			want: "none",
		},
		{
			name: "CompressionGZIP",
			arg:  CompressionGZIP,
			want: "gzip",
		},
		{
			name: "CompressionSnappy",
			arg:  CompressionSnappy,
			want: "snappy",
		},
		{
			name: "CompressionLZ4",
			arg:  CompressionLZ4,
			want: "lz4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.arg != tt.want {
				t.Errorf("%s expected %s, got %s", tt.name, tt.want, tt.arg)
			}
		})
	}
}
