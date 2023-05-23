package consumer

import (
	"reflect"
	"testing"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/constants"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func Test_newMessage(t *testing.T) {
	topic := "t"
	tests := []struct {
		name string
		arg  *kafka.Message
		want *Message
	}{
		{
			name: "default",
			arg: &kafka.Message{
				Key:   []byte("key"),
				Value: []byte("value"),
				TopicPartition: kafka.TopicPartition{
					Offset:    1,
					Partition: 1,
					Topic:     &topic,
				},
				Headers: []kafka.Header{
					{
						Key:   "k",
						Value: []byte("v"),
					},
					{
						Key:   constants.TransactionID,
						Value: []byte("transaction-id"),
					},
				},
			},
			want: &Message{
				Key:               []byte("key"),
				Message:           []byte("value"),
				Offset:            1,
				Partition:         1,
				Topic:             topic,
				PulledDateTimeUTC: time.Now().UTC(),
				transactionID:     "transaction-id",
				headers: map[string]string{
					"k":                     "v",
					constants.TransactionID: "transaction-id",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newMessage(tt.arg)
			got.PulledDateTimeUTC = tt.want.PulledDateTimeUTC
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetHeaders(t *testing.T) {
	tests := []struct {
		name string
		arg  Message
		want map[string]string
	}{
		{
			name: "default",
			arg: Message{
				headers: map[string]string{
					"key": "value",
				},
			},
			want: map[string]string{
				"key": "value",
			},
		},
		{
			name: "nil",
			arg: Message{
				headers: nil,
			},
			want: make(map[string]string),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.arg.GetHeaders(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("m.GetHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetHeader(t *testing.T) {
	type args struct {
		message Message
		key     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default",
			args: args{
				message: Message{
					headers: map[string]string{
						"key": "value",
					},
				},
				key: "key",
			},
			want: "value",
		},
		{
			name: "nil",
			args: args{
				message: Message{
					headers: nil,
				},
				key: "key",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.message.GetHeader(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("m.GetHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetTransactionID(t *testing.T) {
	tests := []struct {
		name string
		arg  Message
		want string
	}{
		{
			name: "default",
			arg: Message{
				transactionID: "1",
			},
			want: "1",
		},
		{
			name: "in header",
			arg: Message{
				transactionID: "",
				headers: map[string]string{
					constants.TransactionID: "1",
				},
			},
			want: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.arg.GetTransactionID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("m.GetTransactionID() = %v, want %v", got, tt.want)
			}
		})
	}
}
