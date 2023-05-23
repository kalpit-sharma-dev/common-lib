package circuit

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
			name: "Default-Config", want: &Config{Enabled: true, TimeoutInSecond: 1, MaxConcurrentRequests: 10,
				ErrorPercentThreshold: 50, RequestVolumeThreshold: 20, SleepWindowInSecond: 5},
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
