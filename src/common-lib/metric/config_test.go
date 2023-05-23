package metric

import (
	"reflect"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/communication/udp"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/util"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		{
			name: "1 Default Configuration Object",
			want: &Config{Communication: udp.New(), Namespace: ""},
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

func TestConfig_GetNamespace(t *testing.T) {
	type fields struct {
		Communication *udp.Config
		Namespace     string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "1 Blank",
			fields: fields{Namespace: ""},
			want:   util.Hostname(util.ProcessName()),
		},
		{
			name:   "2 Test",
			fields: fields{Namespace: "Test"},
			want:   "Test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Communication: tt.fields.Communication,
				Namespace:     tt.fields.Namespace,
			}
			if got := c.GetNamespace(); got != tt.want {
				t.Errorf("Config.GetNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
