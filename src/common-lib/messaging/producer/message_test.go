package producer

import (
	"reflect"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func Test_AddHeader(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args []args
		want map[string]string
	}{
		{
			name: "default",
			args: []args{
				{
					key:   "key",
					value: "value",
				},
			},
			want: map[string]string{
				"key": "value",
			},
		},
		{
			name: "default",
			args: []args{
				{
					key:   "1",
					value: "1",
				},
				{
					key:   "2",
					value: "2",
				},
			},
			want: map[string]string{
				"1": "1",
				"2": "2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Message{}
			for _, arg := range tt.args {
				m.AddHeader(arg.key, arg.value)
			}
			if !reflect.DeepEqual(m.headers, tt.want) {
				t.Errorf("m.headers expected %+v, got %+v", tt.want, m.headers)
			}
		})
	}
}

func Test_toKafkaMessage(t *testing.T) {
	topic := "test"
	tests := []struct {
		name string
		arg  *Message
		want *kafka.Message
	}{
		{
			name: "default",
			arg: &Message{
				Key:   []byte("key"),
				Value: []byte("value"),
				Topic: topic,
				headers: map[string]string{
					"1": "1",
				},
			},
			want: &kafka.Message{
				Key:   []byte("key"),
				Value: []byte("value"),
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: kafka.PartitionAny,
				},
				Headers: []kafka.Header{
					{
						Key:   "1",
						Value: []byte("1"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.arg.toKafkaMessage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("m.toKafkaMessage() = \ngot: %#v, \n\nwant %#v", got, tt.want)
			}
		})
	}
}
