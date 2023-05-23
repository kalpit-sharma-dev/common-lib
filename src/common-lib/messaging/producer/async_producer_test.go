package producer

import (
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func Test_NewAsyncProducer(t *testing.T) {
	tests := []struct {
		name string
		arg  *Config
	}{
		{
			name: "default",
			arg:  NewConfig(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.arg.Address = []string{"localhost"}
			p, err := NewAsyncProducer(tt.arg)
			if err != nil {
				t.Errorf("NewAsyncProducer() errored")
			}
			p.Close()
		})
	}
}

func Test_ProduceChannel(t *testing.T) {
	tests := []struct {
		name string
		arg  *Config
	}{
		{
			name: "default",
			arg:  NewConfig(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.arg.Address = []string{"localhost"}
			p, err := NewAsyncProducer(tt.arg)
			if err != nil {
				t.Errorf("NewAsyncProducer() errored")
			}
			if p.ProduceChannel() == nil {
				t.Errorf("ProduceChannel() returned nil")
			}
			p.ProduceChannel() <- &Message{}
			p.Close()
		})
	}
}

func Test_DeliveryReportChannel(t *testing.T) {
	tests := []struct {
		name string
		arg  *Config
	}{
		{
			name: "default",
			arg:  NewConfig(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.arg.Address = []string{"localhost"}
			p, err := NewAsyncProducer(tt.arg)
			if err != nil {
				t.Errorf("NewAsyncProducer() errored")
			}
			if p.DeliveryReportChannel() == nil {
				t.Errorf("DeliveryReportChannel() returned nil")
			}
			p.Close()
		})
	}
}

func Test_Flush(t *testing.T) {
	tests := []struct {
		name string
		arg  *Config
	}{
		{
			name: "default",
			arg:  NewConfig(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.arg.Address = []string{"localhost"}
			p, err := NewAsyncProducer(tt.arg)
			if err != nil {
				t.Errorf("NewAsyncProducer() errored")
			}
			i := p.Flush(100)
			if i > 0 {
				t.Errorf("Flush should be 0")
			}
			p.ProduceChannel() <- &Message{Topic: "test", Value: []byte("test")}
			<-time.After(5 * time.Millisecond)
			i = p.Flush(100)
			if i != 1 {
				t.Errorf("Flush should be 1, got %d", i)
			}
			p.Close()
		})
	}
}

func Test_Close(t *testing.T) {
	tests := []struct {
		name string
		arg  *Config
	}{
		{
			name: "default",
			arg:  NewConfig(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.arg.Address = []string{"localhost"}
			p, err := NewAsyncProducer(tt.arg)
			if err != nil {
				t.Errorf("NewAsyncProducer() errored")
			}
			p.Close()
			// no error
		})
	}
}

func Test_Health(t *testing.T) {
	tests := []struct {
		name string
		arg  *Config
		want bool
	}{
		{
			name: "default",
			arg:  NewConfig(),
			want: true, // it starts off immediately as true until lib connects to brokers and fails
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.arg.Address = []string{"localhost"}
			p, err := NewAsyncProducer(tt.arg)
			if err != nil {
				t.Errorf("NewAsyncProducer() errored")
			}
			h, err := p.Health()
			if h == nil {
				t.Errorf("Health() returned nil")
			}
			if h.ConnectionState != tt.want {
				t.Errorf("Health().ConnectionState expected %t, got %t", tt.want, h.ConnectionState)
			}
		})
	}
}

func Test_connStateChange(t *testing.T) {
	tests := []struct {
		name string
		arg  bool
		want bool
	}{
		{
			name: "true",
			arg:  true,
			want: true,
		},
		{
			name: "false",
			arg:  false,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			config.Address = []string{"localhost"}
			p, err := NewAsyncProducer(config)
			if err != nil {
				t.Errorf("NewAsyncProducer() errored")
			}
			ap := p.(*asyncProducer)
			ap.connStateChange(tt.arg)
			h, err := p.Health()
			if h == nil {
				t.Errorf("Health() returned nil")
			}
			if h.ConnectionState != tt.want {
				t.Errorf("Health().ConnectionState expected %t, got %t", tt.want, h.ConnectionState)
			}
		})
	}
}

func Test_processMessages(t *testing.T) {
	tests := []struct {
		name string
		arg  []*Message
		want int
	}{
		{
			name: "default",
			arg: []*Message{
				{Topic: "test", Value: []byte("test")},
				{Topic: "test", Value: []byte("test")},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			config.Address = []string{"localhost"}
			p, err := NewAsyncProducer(config)
			if err != nil {
				t.Errorf("NewAsyncProducer() errored")
			}

			ap := p.(*asyncProducer)
			pc := make(chan *Message, len(tt.arg))
			go func() {
				ap.processMessages(pc)
			}()

			for _, m := range tt.arg {
				pc <- m
			}

			// wait to fill producer.ProduceChannel -- cannot wait until after close/done since this channel will be closed as well
			time.Sleep(2 * time.Second)
			// len returns number of messages in rd_kafka_outq_len remove the delivery reports "Events" and "ProductChannel"
			got := ap.producer.Len() - len(ap.producer.ProduceChannel()) + len(ap.producer.Events())

			close(pc)
			if got != tt.want {
				t.Errorf("Expected in queue %d, got %d", tt.want, got)
			}
		})
	}
}

func Test_processEvents(t *testing.T) {
	topic := "test"
	tests := []struct {
		name string
		arg  []kafka.Event
		want int
	}{
		{
			name: "default",
			arg: []kafka.Event{
				&kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &topic}, Value: []byte("test")},
				&kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &topic}, Value: []byte("test")},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			config.Address = []string{"localhost"}
			p, err := NewAsyncProducer(config)
			if err != nil {
				t.Errorf("NewAsyncProducer() errored")
			}
			ap := p.(*asyncProducer)
			ec := make(chan kafka.Event, len(tt.arg))
			done := make(chan bool)
			go func() {
				ap.processEvents(ec)
				close(done)
			}()
			for _, m := range tt.arg {
				ec <- m
			}
			close(ec)
			<-done
			got := len(ap.DeliveryReportChannel())
			if got != tt.want {
				t.Errorf("Expected in drc queue %d, got %d", tt.want, got)
			}
		})
	}
}
