package publisher

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Shopify/sarama"
)

func Test_syncProducer_Publish(t *testing.T) {
	t.Run("No Messages", func(t *testing.T) {
		s := &syncProducer{cfg: NewConfig()}
		if err := s.Publish(context.Background(), "transaction"); err == nil {
			t.Errorf("syncProducer.Publish() want error got nil")
		}
	})

	t.Run("One Message", func(t *testing.T) {
		s := &syncProducer{cfg: NewConfig()}
		if err := s.Publish(context.Background(), "transaction", &Message{}); err == nil {
			t.Errorf("syncProducer.Publish() want error got nil")
		}
	})

	t.Run("Blank Valid Errors", func(t *testing.T) {
		cfg := NewConfig()
		cfg.CircuitProneErrors = []string{}
		cfg.Address = []string{"localhost"}
		s := &syncProducer{cfg: cfg}
		if err := s.Publish(context.Background(), "transaction", &Message{}); err == nil {
			t.Errorf("syncProducer.Publish() want error got nil")
		}
	})

	t.Run("Invalid Address Error", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Address = []string{"localhost"}
		s := &syncProducer{cfg: cfg}
		if err := s.Publish(context.Background(), "transaction", &Message{Topic: "test"}); err == nil {
			t.Errorf("syncProducer.Publish() want error got nil")
		}
	})
}

func Test_syncProducer_publish(t *testing.T) {
	cntx, cancle := context.WithTimeout(context.Background(), time.Nanosecond)
	cancle()
	type fields struct {
		cfg          *Config
		producerType ProducerType
		producer     sarama.SyncProducer
		existing     sarama.SyncProducer
	}
	type args struct {
		ctx         context.Context
		transaction string
		messages    []*Message
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "1", args: args{ctx: context.Background()}, wantErr: true},
		{
			name: "2", args: args{ctx: context.Background()}, wantErr: true,
			fields: fields{cfg: &Config{}, producerType: BigKafkaProducer},
		},
		{
			name: "3", args: args{ctx: context.Background()}, wantErr: false,
			fields: fields{cfg: &Config{Address: []string{"localhost"}}, producerType: BigKafkaProducer},
		},
		{
			name: "4", args: args{ctx: context.Background(), messages: []*Message{&Message{}}}, wantErr: true,
			fields: fields{cfg: &Config{Address: []string{"localhost"}}, producerType: BigKafkaProducer},
		},
		{
			name: "5", args: args{ctx: cntx, messages: []*Message{&Message{}}}, wantErr: true,
			fields: fields{cfg: &Config{Address: []string{"localhost"}}, producerType: BigKafkaProducer},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncProducer{
				cfg:          tt.fields.cfg,
				producerType: tt.fields.producerType,
				producer:     tt.fields.producer,
				existing:     tt.fields.existing,
			}
			if err := s.publish(tt.args.ctx, tt.args.transaction, tt.args.messages...); (err != nil) != tt.wantErr {
				t.Errorf("syncProducer.publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_syncProducer_publish_concurrently_avoid_datarace(t *testing.T) {
	cfg := NewConfig()
	cfg.Address = []string{"non-existing-address"}

	producer, err := SyncProducer(RegularKafkaProducer, cfg)
	if err != nil {
		t.Errorf("failed to create producer: %+v", err)
	}

	wg := &sync.WaitGroup{}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := &Message{
				Topic: "topic",
				Key:   EncodeString("key"),
				Value: EncodeString("value"),
			}
			_ = producer.Publish(context.Background(), "txID", msg)
		}()
	}
	wg.Wait()
}

func TestGetSaramaErrorMsg(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "error_with_sarama_errors",
			args: args{
				err: sarama.ProducerErrors{&sarama.ProducerError{
					Msg: &sarama.ProducerMessage{Topic: "companysite-entities-v0"},
					Err: sarama.ErrOutOfBrokers,
				}},
			},
			want: `kafka: Failed to produce message to topic companysite-entities-v0: kafka: client has run out of available brokers to talk to (Is your cluster reachable?)\n`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSaramaErrorMsg(tt.args.err); strings.Contains(got, tt.want) {
				t.Errorf("GetSaramaErrorMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}
