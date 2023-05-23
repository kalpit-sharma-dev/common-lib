package udp

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		{
			name: "1. Default Configuration",
			want: &Config{Address: "localhost", PortNumber: "7000", TimeoutInSeconds: 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_ServerAddress(t *testing.T) {
	type fields struct {
		Address          string
		PortNumber       string
		TimeoutInSeconds int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "1. Local Address",
			fields: fields{Address: "localhost", PortNumber: "7000"},
			want:   "localhost:7000",
		},
		{
			name:   "2. IP Address",
			fields: fields{Address: "1.1.1.1", PortNumber: "6000"},
			want:   "1.1.1.1:6000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Address:          tt.fields.Address,
				PortNumber:       tt.fields.PortNumber,
				TimeoutInSeconds: tt.fields.TimeoutInSeconds,
			}
			if got := c.ServerAddress(); got != tt.want {
				t.Errorf("Config.Address() = %v, want %v", got, tt.want)
			}
		})
	}
}
