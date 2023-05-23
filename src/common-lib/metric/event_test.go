package metric

import (
	"reflect"
	"testing"
)

func TestCreateEvent(t *testing.T) {
	type args struct {
		title       string
		description string
	}
	tests := []struct {
		name string
		args args
		want *Event
	}{
		{
			name: "1 Event-1",
			args: args{title: "EVENT_ERROR", description: "Test Error"},
			want: &Event{Title: "EVENT_ERROR", Description: "Test Error", Properties: map[string]string{}},
		},
		{
			name: "2 Event-2",
			args: args{title: "AVAILABILITY_EVENT", description: "this is availability"},
			want: &Event{Title: "AVAILABILITY_EVENT", Description: "this is availability", Properties: map[string]string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateEvent(tt.args.title, tt.args.description); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}
