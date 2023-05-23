package consumer

import (
	"reflect"
	"testing"

	"github.com/Shopify/sarama"
)

func Test_getConsumerStrategy(t *testing.T) {
	t.Run("1 pullSequential", func(t *testing.T) {
		got := getConsumerStrategy(Config{ConsumerMode: PullOrdered})
		_, ok := got.(*pullOrdered)
		if !ok {
			t.Errorf("getConsumerStrategy() = %v, want %v", got, &pullOrdered{})
		}
	})

	t.Run("2 pullParallel", func(t *testing.T) {
		got := getConsumerStrategy(Config{ConsumerMode: PullUnOrdered})
		_, ok := got.(*pullUnOrdered)
		if !ok {
			t.Errorf("getConsumerStrategy() = %v, want %v", got, &pullUnOrdered{})
		}
	})

	t.Run("3 pullSequentialWithOffset", func(t *testing.T) {
		got := getConsumerStrategy(Config{ConsumerMode: PullOrderedWithOffsetReplay})
		_, ok := got.(*pullOrderedWithOffsetReplay)
		if !ok {
			t.Errorf("getConsumerStrategy() = %v, want %v", got, &pullOrderedWithOffsetReplay{})
		}
	})

	t.Run("4 Unknown", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		got := getConsumerStrategy(Config{ConsumerMode: 0})
		t.Errorf("getConsumerStrategy() = %v, want %v", got, "Panic")
	})

	t.Run("5 Unknown", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		got := getConsumerStrategy(Config{ConsumerMode: -4})
		t.Errorf("getConsumerStrategy() = %v, want %v", got, "Panic")
	})

}

func Test_newMessage(t *testing.T) {
	type args struct {
		message *sarama.ConsumerMessage
	}
	tests := []struct {
		name string
		args args
		want Message
	}{
		{
			name: "Without Header", args: args{message: &sarama.ConsumerMessage{
				Key: []byte("Key"), Value: []byte("Value"), Topic: "test", Partition: 1, Offset: 2}},
			want: Message{Message: []byte("Value"), Topic: "test", Partition: 1, Offset: 2, headers: map[string]string{}},
		},
		{
			name: "With Header", args: args{message: &sarama.ConsumerMessage{
				Key: []byte("Key"), Value: []byte("Value"), Topic: "test", Partition: 1, Offset: 2,
				Headers: []*sarama.RecordHeader{&sarama.RecordHeader{Key: []byte("Key"), Value: []byte("Value")}},
			}},
			want: Message{Message: []byte("Value"), Topic: "test", Partition: 1, Offset: 2,
				headers: map[string]string{"Key": "Value"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newMessage(tt.args.message)
			tt.want.PulledDateTimeUTC = got.PulledDateTimeUTC
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
