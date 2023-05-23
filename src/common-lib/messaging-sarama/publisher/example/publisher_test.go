package main

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging-sarama/publisher"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging-sarama/publisher/mock"
)

func Test_produce(t *testing.T) {
	ctrl := gomock.NewController(t)
	producer := mock.NewMockProducer(ctrl)
	type args struct {
		cfg         *Config
		transaction string
		count       int
		index       int
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "ProducerError", wantErr: true, args: args{cfg: &Config{Pub: &publisher.Config{}}},
			setup: func() {
				publisher.SyncProducer = func(producerType publisher.ProducerType, cfg *publisher.Config) (publisher.Producer, error) {
					return nil, errors.New("Error")
				}
			},
		},
		{
			name: "Publish-Error", wantErr: true, args: args{cfg: &Config{Pub: &publisher.Config{}}},
			setup: func() {
				publisher.SyncProducer = func(producerType publisher.ProducerType, cfg *publisher.Config) (publisher.Producer, error) {
					return producer, nil
				}
				producer.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("Error"))
			},
		},
		{
			name: "Publish-Success", wantErr: false, args: args{cfg: &Config{Topics: []string{"test"}, Pub: &publisher.Config{}}},
			setup: func() {
				publisher.SyncProducer = func(producerType publisher.ProducerType, cfg *publisher.Config) (publisher.Producer, error) {
					return producer, nil
				}
				producer.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			if err := produce(tt.args.cfg, tt.args.transaction, tt.args.count, tt.args.index); (err != nil) != tt.wantErr {
				t.Errorf("produce() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_main(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "ProducerError",
			setup: func() {
				publisher.SyncProducer = func(producerType publisher.ProducerType, cfg *publisher.Config) (publisher.Producer, error) {
					return nil, errors.New("Error")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			main()
		})
	}
}
