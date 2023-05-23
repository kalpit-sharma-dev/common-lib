package tracing

import (
	"reflect"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
)

func TestNewConfigDefaults(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		{name: "default", want: &Config{
			Type:           AwsXrayTracingType,
			HostPlatform:   "",
			Enabled:        true,
			ServiceName:    utils.GetServiceName(),
			ServiceVersion: utils.GetServiceVersion(),
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
