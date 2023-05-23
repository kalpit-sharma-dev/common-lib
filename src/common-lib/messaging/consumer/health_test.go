package consumer

import (
	"reflect"
	"testing"
)

func Test_newHealth(t *testing.T) {
	tests := []struct {
		name string
		want *Health
	}{
		{
			name: "default",
			want: &Health{
				ConnectionState: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newHealth(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newHealth() = %v, want %v", got, tt.want)
			}
		})
	}
}
