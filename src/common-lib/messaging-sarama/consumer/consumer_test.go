package consumer

import (
	"reflect"
	"testing"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/constants"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
		want Config
	}{
		{name: "default", want: Config{
			SubscriberPerCore: 20,
			CommitMode:        OnPull,
			ConsumerMode:      PullUnOrdered,
			Retention:         0,
			OffsetsInitial:    OffsetNewest,
			Timeout:           time.Minute,
			MetadataFull:      true,
			RetryCount:        10,
			RetryDelay:        30 * time.Second,
			RebalanceTimeout:  60 * time.Second,
			Partitions:        500,
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

func Test_newHealth(t *testing.T) {
	type args struct {
		cfg Config
	}
	tests := []struct {
		name string
		args args
		want *Health
	}{
		{name: "1 Blank Config", args: args{cfg: Config{}}, want: &Health{CanCosume: false,
			ConnectionState: make(map[string]bool), PerTopicPartitions: make(map[string][]int32)}},
		{name: "2 Single IP", args: args{cfg: Config{Address: []string{"1"}}}, want: &Health{CanCosume: false,
			PerTopicPartitions: make(map[string][]int32), ConnectionState: map[string]bool{"1": false}}},
		{name: "3 Multiple IP", args: args{cfg: Config{Address: []string{"1", "2"}, Topics: []string{"test"}, Group: "2"}},
			want: &Health{CanCosume: false, PerTopicPartitions: make(map[string][]int32),
				Topics: []string{"test"}, Group: "2",
				ConnectionState: map[string]bool{"1": false, "2": false}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newHealth(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newHealth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_GetHeader(t *testing.T) {
	type fields struct {
		Message           []byte
		Offset            int64
		Partition         int32
		Topic             string
		PulledDateTimeUTC time.Time
		headers           map[string]string
		transactionID     string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{name: "Invalid", args: args{key: "Invalid"}, want: ""},
		{name: "Invalid Test", fields: fields{headers: map[string]string{"Test": "Test"}}, args: args{key: "Invalid Test"}, want: ""},
		{name: "Test", fields: fields{headers: map[string]string{"Test": "Test"}}, args: args{key: "Test"}, want: "Test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Message:           tt.fields.Message,
				Offset:            tt.fields.Offset,
				Partition:         tt.fields.Partition,
				Topic:             tt.fields.Topic,
				PulledDateTimeUTC: tt.fields.PulledDateTimeUTC,
				headers:           tt.fields.headers,
				transactionID:     tt.fields.transactionID,
			}
			if got := m.GetHeader(tt.args.key); got != tt.want {
				t.Errorf("Message.GetHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_GetTransactionID(t *testing.T) {
	type fields struct {
		Message           []byte
		Offset            int64
		Partition         int32
		Topic             string
		PulledDateTimeUTC time.Time
		headers           map[string]string
		transactionID     string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Transaction ID On Message", want: "TransactionID",
			fields: fields{transactionID: "TransactionID", headers: map[string]string{constants.TransactionID: "TransactionID"}},
		},
		{
			name: "Transaction ID not on Message", want: "TransactionID",
			fields: fields{headers: map[string]string{constants.TransactionID: "TransactionID"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Message:           tt.fields.Message,
				Offset:            tt.fields.Offset,
				Partition:         tt.fields.Partition,
				Topic:             tt.fields.Topic,
				PulledDateTimeUTC: tt.fields.PulledDateTimeUTC,
				headers:           tt.fields.headers,
				transactionID:     tt.fields.transactionID,
			}
			if got := m.GetTransactionID(); got != tt.want {
				t.Errorf("Message.GetTransactionID() = %v, want %v", got, tt.want)
			}
		})
	}
}
