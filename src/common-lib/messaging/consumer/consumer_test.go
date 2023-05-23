package consumer

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gammazero/workerpool"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestKafkaConsumer_Close(t *testing.T) {
	topic := "topic"
	type fields struct {
		closing          chan bool
		pausedPartitions map[partition]kafka.TopicPartition
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Good",
			fields: fields{
				closing: make(chan bool, 1),
				pausedPartitions: map[partition]kafka.TopicPartition{
					{
						topic:     topic,
						partition: 0,
					}: {
						Topic:     &topic,
						Partition: 0,
						Offset:    0,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &KafkaConsumer{
				closing: tt.fields.closing,

				partMx:           &sync.Mutex{},
				pausedPartitions: tt.fields.pausedPartitions,
			}

			err := c.Close()

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			cl, ok := <-c.closing
			require.True(t, ok)
			require.False(t, cl)

			_, ok = <-c.closing
			require.False(t, ok)

			require.Empty(t, c.pausedPartitions)
		})
	}
}

func TestKafkaConsumer_CloseWait(t *testing.T) {
	topic := "topic"
	type fields struct {
		closing          chan bool
		partMx           *sync.Mutex
		pausedPartitions map[partition]kafka.TopicPartition
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Good",
			fields: fields{
				closing: make(chan bool, 1),
				partMx:  &sync.Mutex{},
				pausedPartitions: map[partition]kafka.TopicPartition{
					{
						topic:     topic,
						partition: 0,
					}: {
						Topic:     &topic,
						Partition: 0,
						Offset:    0,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &KafkaConsumer{
				closing: tt.fields.closing,

				partMx:           tt.fields.partMx,
				pausedPartitions: tt.fields.pausedPartitions,
			}

			err := c.CloseWait()

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			cl, ok := <-c.closing
			require.True(t, ok)
			require.True(t, cl)

			_, ok = <-c.closing
			require.False(t, ok)

			require.Empty(t, c.pausedPartitions)
		})
	}
}

func TestKafkaConsumer_Health(t *testing.T) {
	health := Health{
		ConnectionState: false,
		Address:         []string{"127.0.0.1"},
		Topics:          []string{"aa"},
		Group:           "group",
	}
	t.Run("Get Health", func(t *testing.T) {
		c := &KafkaConsumer{
			healthMx: &sync.RWMutex{},
			health:   &health,
		}
		got, err := c.Health()

		require.NoError(t, err)
		require.Equal(t, health, got)
	})
}

func TestKafkaConsumer_MarkOffset(t *testing.T) {
	type fields struct {
		consumerFactory func(t *testing.T, ctrl *gomock.Controller) consumer
		loggerFactory   func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger)
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Good",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().StoreOffsets(gomock.Any()).Return(nil, nil).Times(1) // actual test
					return c
				},
				loggerFactory: func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger) {
					t.Helper()
					log := NewMockConsumerLogger(ctrl)
					return log, log, log
				},
			},
		},
		{
			name: "Errors are logged",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().StoreOffsets(gomock.Any()).Return(nil, errors.New("oops")).Times(1) // actual test
					return c
				},
				loggerFactory: func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger) {
					t.Helper()
					log := NewMockConsumerLogger(ctrl)
					log.EXPECT().Printf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
					return log, log, log
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			er, in, de := tt.fields.loggerFactory(t, ctrl)
			c := &KafkaConsumer{
				consumer: tt.fields.consumerFactory(t, ctrl),
				errorLog: er,
				infoLog:  in,
				debugLog: de,
			}
			c.MarkOffset("topic", 0, 100)
		})
	}
}

func TestKafkaConsumer_Pause(t *testing.T) {
	topic := "topic"
	type fields struct {
		consumerFactory  func(t *testing.T, ctrl *gomock.Controller) consumer
		pausedPartitions map[partition]kafka.TopicPartition
	}
	type args struct {
		topic  string
		part   int32
		offset int64
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		wantPausedPartitions map[partition]kafka.TopicPartition
		wantErr              bool
	}{
		{
			name: "Already paused",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{
					{
						topic:     topic,
						partition: 0,
					}: {
						Topic:     &topic,
						Partition: 0,
						Offset:    1,
					},
				},
			},
			args: args{
				topic:  topic,
				part:   0,
				offset: 1,
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{
				{
					topic:     topic,
					partition: 0,
				}: {
					Topic:     &topic,
					Partition: 0,
					Offset:    1,
				},
			},
			wantErr: false,
		},
		{
			name: "Good",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Pause([]kafka.TopicPartition{
						{
							Topic:     &topic,
							Partition: 0,
							Offset:    1,
						},
					}).Return(nil).Times(1)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{},
			},
			args: args{
				topic:  topic,
				part:   0,
				offset: 1,
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{
				{
					topic:     topic,
					partition: 0,
				}: {
					Topic:     &topic,
					Partition: 0,
					Offset:    1,
				},
			},
			wantErr: false,
		},
		{
			name: "Pause returns error",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Pause(gomock.Any()).Return(errors.New("oops")).Times(1)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{},
			},
			args: args{
				topic:  topic,
				part:   0,
				offset: 1,
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{},
			wantErr:              true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := &KafkaConsumer{
				consumer: tt.fields.consumerFactory(t, ctrl),

				partMx:           &sync.Mutex{},
				pausedPartitions: tt.fields.pausedPartitions,
			}

			err := c.Pause(tt.args.topic, tt.args.part, tt.args.offset)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantPausedPartitions, c.pausedPartitions)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.wantPausedPartitions, c.pausedPartitions)
		})
	}
}

func TestKafkaConsumer_PauseAll(t *testing.T) {
	topic := "topic"
	type fields struct {
		consumerFactory  func(t *testing.T, ctrl *gomock.Controller) consumer
		pausedPartitions map[partition]kafka.TopicPartition
	}
	tests := []struct {
		name                 string
		fields               fields
		wantPausedPartitions map[partition]kafka.TopicPartition
		wantErr              bool
	}{
		{
			name: "Some are Already paused",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Assignment().Return([]kafka.TopicPartition{
						{
							Topic:     &topic,
							Partition: 0,
							Offset:    1,
						},
						{
							Topic:     &topic,
							Partition: 1,
							Offset:    1,
						},
						{
							Topic:     &topic,
							Partition: 2,
							Offset:    1,
						},
					}, nil).Times(1)
					c.EXPECT().Pause([]kafka.TopicPartition{
						{
							Topic:     &topic,
							Partition: 1,
							Offset:    1,
						},
						{
							Topic:     &topic,
							Partition: 2,
							Offset:    1,
						},
					}).Return(nil).Times(1)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{
					{
						topic:     topic,
						partition: 0,
					}: {
						Topic:     &topic,
						Partition: 0,
						Offset:    1,
					},
				},
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{
				{
					topic:     topic,
					partition: 0,
				}: {
					Topic:     &topic,
					Partition: 0,
					Offset:    1,
				},
				{
					topic:     topic,
					partition: 1,
				}: {
					Topic:     &topic,
					Partition: 1,
					Offset:    1,
				},
				{
					topic:     topic,
					partition: 2,
				}: {
					Topic:     &topic,
					Partition: 2,
					Offset:    1,
				},
			},
			wantErr: false,
		},
		{
			name: "Pause returns error",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Assignment().Return([]kafka.TopicPartition{
						{
							Topic:     &topic,
							Partition: 0,
							Offset:    1,
						},
					}, nil).Times(1)
					c.EXPECT().Pause(gomock.Any()).Return(errors.New("oops")).Times(1)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{},
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{},
			wantErr:              true,
		},
		{
			name: "Assignment returns error",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Assignment().Return([]kafka.TopicPartition{
						{
							Topic:     &topic,
							Partition: 0,
							Offset:    1,
						},
					}, errors.New("oops")).Times(1)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{},
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{},
			wantErr:              true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := &KafkaConsumer{
				consumer:         tt.fields.consumerFactory(t, ctrl),
				partMx:           &sync.Mutex{},
				pausedPartitions: tt.fields.pausedPartitions,
			}

			err := c.PauseAll()

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantPausedPartitions, c.pausedPartitions)
				return
			}
			require.Equal(t, tt.wantPausedPartitions, c.pausedPartitions)
			require.NoError(t, err)
		})
	}
}

func TestKafkaConsumer_Pull_closingConsumer(t *testing.T) {
	type fields struct {
		config                  *Config
		consumerFactory         func(t *testing.T, ctrl *gomock.Controller) consumer
		consumerStrategyFactory func(t *testing.T, ctrl *gomock.Controller) consumerStrategy
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Good",
			fields: fields{
				config: &Config{TransactionID: "1"},
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Close().Return(nil).Times(1) // actual test
					return c
				},
				consumerStrategyFactory: func(t *testing.T, ctrl *gomock.Controller) consumerStrategy {
					t.Helper()
					c := NewMockconsumerStrategy(ctrl)
					c.EXPECT().close(gomock.Any()).Times(1) // actual test
					return c
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			log := NewMockConsumerLogger(ctrl)
			log.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

			closing := make(chan bool, 1)
			closing <- false
			close(closing)

			c := &KafkaConsumer{
				config:           tt.fields.config,
				closing:          closing,
				consumer:         tt.fields.consumerFactory(t, ctrl),
				consumerStrategy: tt.fields.consumerStrategyFactory(t, ctrl),
				infoLog:          log,
			}

			c.Pull() // wait for return
		})
	}
}

func TestKafkaConsumer_Pull_handleMessage(t *testing.T) {
	startegyClosed := sync.WaitGroup{}

	topic := "topic"
	notificationHandlerWG := sync.WaitGroup{}
	notificationHandlerWG.Add(1)
	config := &Config{
		ErrorHandler: func(context.Context, error, *Message) {
			notificationHandlerWG.Done()
		},
		TransactionID: "1",
	}
	health := &Health{ConnectionState: false}

	consumerFactory := func(t *testing.T, ctrl *gomock.Controller) consumer {
		t.Helper()
		c := NewMockconsumer(ctrl)
		call1 := c.EXPECT().Poll(gomock.Any()).Return(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: 0,
				Offset:    1,
			},
		}).Times(1)
		call2 := c.EXPECT().Poll(gomock.Any()).DoAndReturn(func(int) kafka.Event {
			time.Sleep(10 * time.Millisecond)
			return nil
		}).AnyTimes()
		gomock.InOrder(call1, call2)
		c.EXPECT().Close().Return(nil).AnyTimes()
		return c
	}
	consumerStrategyFactory := func(t *testing.T, ctrl *gomock.Controller) consumerStrategy {
		t.Helper()
		c := NewMockconsumerStrategy(ctrl)
		notificationHandlerWG.Add(1)
		c.EXPECT().handleMessage(gomock.Any()).DoAndReturn(func(p *kafka.Message) error {
			notificationHandlerWG.Done()
			return errors.New("oops")
		},
		).Times(1)
		startegyClosed.Add(1)
		c.EXPECT().close(gomock.Any()).Do(func(_ bool) {
			startegyClosed.Done()
		}).Times(1)
		return c
	}

	t.Run("message processing error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := NewMockConsumerLogger(ctrl)
		log.EXPECT().Printf(gomock.Any(), gomock.Any()).AnyTimes()

		closing := make(chan bool, 1)

		c := &KafkaConsumer{
			config:           config,
			closing:          closing,
			consumer:         consumerFactory(t, ctrl),
			consumerStrategy: consumerStrategyFactory(t, ctrl),
			infoLog:          log,
			health:           health,
		}

		go c.Pull()

		waitOrFail(t, &notificationHandlerWG, 5*time.Second)
		err := c.close(false)
		require.NoError(t, err)
		startegyClosed.Wait()
	})
}

func TestKafkaConsumer_Resume(t *testing.T) {
	topic := "topic"
	type fields struct {
		consumerFactory  func(t *testing.T, ctrl *gomock.Controller) consumer
		pausedPartitions map[partition]kafka.TopicPartition
	}
	type args struct {
		topic  string
		part   int32
		offset int64
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		wantPausedPartitions map[partition]kafka.TopicPartition
		wantErr              bool
	}{
		{
			name: "Not paused",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{},
			},
			args: args{
				topic:  topic,
				part:   0,
				offset: 1,
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{},
			wantErr:              false,
		},
		{
			name: "Good",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Resume([]kafka.TopicPartition{
						{
							Topic:     &topic,
							Partition: 0,
							Offset:    1,
						},
					}).Return(nil).Times(1)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{
					{
						topic:     topic,
						partition: 0,
					}: {
						Topic:     &topic,
						Partition: 0,
						Offset:    1,
					},
				},
			},
			args: args{
				topic:  topic,
				part:   0,
				offset: 1,
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{},
			wantErr:              false,
		},
		{
			name: "Resume returns error",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Resume(gomock.Any()).Return(errors.New("oops")).Times(1)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{
					{
						topic:     topic,
						partition: 0,
					}: {
						Topic:     &topic,
						Partition: 0,
						Offset:    1,
					},
				},
			},
			args: args{
				topic:  topic,
				part:   0,
				offset: 1,
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{
				{
					topic:     topic,
					partition: 0,
				}: {
					Topic:     &topic,
					Partition: 0,
					Offset:    1,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := &KafkaConsumer{
				consumer: tt.fields.consumerFactory(t, ctrl),

				partMx:           &sync.Mutex{},
				pausedPartitions: tt.fields.pausedPartitions,
			}

			err := c.Resume(tt.args.topic, tt.args.part, tt.args.offset)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantPausedPartitions, c.pausedPartitions)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantPausedPartitions, c.pausedPartitions)
		})
	}
}

func TestKafkaConsumer_ResumeAll(t *testing.T) {
	topic := "topic"
	type fields struct {
		consumerFactory  func(t *testing.T, ctrl *gomock.Controller) consumer
		pausedPartitions map[partition]kafka.TopicPartition
	}
	tests := []struct {
		name                 string
		fields               fields
		wantPausedPartitions map[partition]kafka.TopicPartition
		wantErr              bool
	}{
		{
			name: "Good",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Resume(gomock.Any()).Return(nil).Times(1)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{
					{
						topic:     topic,
						partition: 1,
					}: {
						Topic:     &topic,
						Partition: 1,
						Offset:    1,
					},
					{
						topic:     topic,
						partition: 2,
					}: {
						Topic:     &topic,
						Partition: 2,
						Offset:    1,
					},
				},
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{},
			wantErr:              false,
		},
		{
			name: "Nothing is paused",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{},
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{},
			wantErr:              false,
		},
		{
			name: "Pause returns error",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Resume(gomock.Any()).Return(errors.New("oops")).Times(1)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{
					{
						topic:     topic,
						partition: 0,
					}: {
						Topic:     &topic,
						Partition: 0,
						Offset:    1,
					},
				},
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{
				{
					topic:     topic,
					partition: 0,
				}: {
					Topic:     &topic,
					Partition: 0,
					Offset:    1,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := &KafkaConsumer{
				consumer:         tt.fields.consumerFactory(t, ctrl),
				partMx:           &sync.Mutex{},
				pausedPartitions: tt.fields.pausedPartitions,
			}

			err := c.ResumeAll()

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantPausedPartitions, c.pausedPartitions)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantPausedPartitions, c.pausedPartitions)
		})
	}
}

func TestKafkaConsumer_close(t *testing.T) {
	type args struct {
		wait bool
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Wait true",
			args: args{
				wait: true,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Wait false",
			args: args{
				wait: false,
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &KafkaConsumer{
				closing: make(chan bool, 1),
			}

			err := c.close(tt.args.wait)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			got, ok := <-c.closing
			require.True(t, ok)
			require.Equal(t, tt.want, got)

			_, ok = <-c.closing
			require.False(t, ok)
		})
	}
}

func TestKafkaConsumer_handleAssignedPartitions(t *testing.T) {
	topic := "topic"
	type fields struct {
		consumerFactory         func(t *testing.T, ctrl *gomock.Controller) consumer
		consumerStrategyFactory func(t *testing.T, ctrl *gomock.Controller) consumerStrategy
		health                  *Health
		pausedPartitions        map[partition]kafka.TopicPartition
	}
	type args struct {
		p kafka.AssignedPartitions
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		wantPausedPartitions map[partition]kafka.TopicPartition
		wantHealth           *Health
	}{
		{
			name: "Good",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Resume(gomock.Any()).Return(nil).Times(1)
					c.EXPECT().Assign(gomock.Any()).Return(nil).Times(1)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{
					{
						topic:     topic,
						partition: 0,
					}: {
						Topic:     &topic,
						Partition: 0,
						Offset:    1,
					},
				},
				consumerStrategyFactory: func(t *testing.T, ctrl *gomock.Controller) consumerStrategy {
					t.Helper()
					c := NewMockconsumerStrategy(ctrl)
					c.EXPECT().handleAssignedPartitions(gomock.Any(), gomock.Any()).Times(1)
					return c
				},
				health: newHealth(),
			},
			args: args{
				p: kafka.AssignedPartitions{
					Partitions: kafka.TopicPartitions{
						{
							Topic:     &topic,
							Partition: 0,
							Offset:    1,
						},
					},
				},
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{},
			wantHealth:           &Health{ConnectionState: true},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := &KafkaConsumer{
				consumer:         tt.fields.consumerFactory(t, ctrl),
				consumerStrategy: tt.fields.consumerStrategyFactory(t, ctrl),
				healthMx:         &sync.RWMutex{},
				health:           tt.fields.health,
				partMx:           &sync.Mutex{},
				pausedPartitions: tt.fields.pausedPartitions,
			}

			c.handleAssignedPartitions(tt.args.p)

			require.Equal(t, tt.wantPausedPartitions, c.pausedPartitions)
		})
	}
}

func TestKafkaConsumer_handleRevokedPartitions(t *testing.T) {
	topic := "topic"
	type fields struct {
		consumerFactory         func(t *testing.T, ctrl *gomock.Controller) consumer
		consumerStrategyFactory func(t *testing.T, ctrl *gomock.Controller) consumerStrategy
		pausedPartitions        map[partition]kafka.TopicPartition
	}
	type args struct {
		p kafka.RevokedPartitions
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		wantPausedPartitions map[partition]kafka.TopicPartition
	}{
		{
			name: "Good",
			fields: fields{
				consumerFactory: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Resume(gomock.Any()).Return(nil).Times(1)
					c.EXPECT().Unassign().Return(nil).Times(1)
					return c
				},
				pausedPartitions: map[partition]kafka.TopicPartition{
					{
						topic:     topic,
						partition: 0,
					}: {
						Topic:     &topic,
						Partition: 0,
						Offset:    1,
					},
				},
				consumerStrategyFactory: func(t *testing.T, ctrl *gomock.Controller) consumerStrategy {
					t.Helper()
					c := NewMockconsumerStrategy(ctrl)
					c.EXPECT().handleRevokedPartitions(gomock.Any()).Times(1)
					return c
				},
			},
			args: args{
				p: kafka.RevokedPartitions{
					Partitions: kafka.TopicPartitions{
						{
							Topic:     &topic,
							Partition: 0,
							Offset:    1,
						},
					},
				},
			},
			wantPausedPartitions: map[partition]kafka.TopicPartition{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := &KafkaConsumer{
				consumer:         tt.fields.consumerFactory(t, ctrl),
				consumerStrategy: tt.fields.consumerStrategyFactory(t, ctrl),
				partMx:           &sync.Mutex{},
				pausedPartitions: tt.fields.pausedPartitions,
			}

			c.handleRevokedPartitions(tt.args.p)
		})
	}
}

func TestKafkaConsumer_process(t *testing.T) {
	type fields struct {
		config                *Config
		commitStrategyFactory func(t *testing.T, ctrl *gomock.Controller) commitStrategy
	}
	type args struct {
		message *Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Good",
			fields: fields{
				config: &Config{
					RetryCount: 1,
					Timeout:    5 * time.Second,
					MessageHandler: func(ctx context.Context, message Message) error {
						return nil
					},
				},
				commitStrategyFactory: func(t *testing.T, ctrl *gomock.Controller) commitStrategy {
					t.Helper()
					cs := NewMockcommitStrategy(ctrl)
					cs.EXPECT().onPull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
					cs.EXPECT().beforeHandler(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
					cs.EXPECT().afterHandler(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
					return cs
				},
			},
			args: args{
				message: &Message{
					Offset:        0,
					Partition:     0,
					Topic:         "topic",
					transactionID: "1",
				},
			},
		},
		{
			name: "Failed all retries",
			fields: fields{
				config: &Config{
					RetryCount: 2,
					Timeout:    5 * time.Second,
					MessageHandler: func(ctx context.Context, message Message) error {
						return errors.New("oops")
					},
				},
				commitStrategyFactory: func(t *testing.T, ctrl *gomock.Controller) commitStrategy {
					t.Helper()
					cs := NewMockcommitStrategy(ctrl)
					cs.EXPECT().onPull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
					cs.EXPECT().beforeHandler(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
					cs.EXPECT().afterHandler(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
					return cs
				},
			},
			args: args{
				message: &Message{
					Offset:        0,
					Partition:     0,
					Topic:         "topic",
					transactionID: "1",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			log := NewMockConsumerLogger(ctrl)
			log.EXPECT().Printf(gomock.Any(), gomock.Any()).AnyTimes()

			c := &KafkaConsumer{
				config:         tt.fields.config,
				commitStrategy: tt.fields.commitStrategyFactory(t, ctrl),
				debugLog:       log,
			}
			c.process(tt.args.message)
		})
	}
}

func Test_getResetPosition(t *testing.T) {
	type args struct {
		config *Config
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "OffsetNewest",
			args: args{
				config: &Config{OffsetsInitial: OffsetNewest},
			},
			want: "latest",
		},
		{
			name: "OffsetOldest",
			args: args{
				config: &Config{OffsetsInitial: OffsetOldest},
			},
			want: "earliest",
		},
		{
			name: "Default",
			args: args{
				config: &Config{},
			},
			want: "latest",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			got := getResetPosition(tt.args.config)

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_invokeErrorHandler(t *testing.T) {
	type fields struct {
		config        *Config
		loggerFactory func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger)
	}
	type args struct {
		err     error
		message *Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Good w/o message and handler",
			fields: fields{
				config: NewConfig(),
				loggerFactory: func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger) {
					t.Helper()
					log := NewMockConsumerLogger(ctrl)
					log.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)
					return log, log, log
				},
			},
			args: args{
				err:     errors.New("oops"),
				message: nil,
			},
		},
		{
			name: "Handler panics",
			fields: fields{
				config: &Config{
					ErrorHandler: func(ctx context.Context, err error, message *Message) {
						panic(errors.New("oops"))
					},
					Timeout:              2 * time.Second,
					ErrorHandlingTimeout: 2 * time.Second,
				},
				loggerFactory: func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger) {
					t.Helper()
					log := NewMockConsumerLogger(ctrl)
					log.EXPECT().Printf(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
					log.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)
					return log, log, log
				},
			},
			args: args{
				err:     errors.New("oops"),
				message: nil,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			er, _, de := tt.fields.loggerFactory(t, ctrl)
			c := &KafkaConsumer{
				config:   tt.fields.config,
				errorLog: er,
				debugLog: de,
			}

			c.invokeErrorHandler(tt.args.err, tt.args.message)
		})
	}
}

func Test_invokeMessageHandler(t *testing.T) {
	type fields struct {
		config        *Config
		loggerFactory func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger)
	}
	type args struct {
		message *Message
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Good",
			fields: fields{
				config: &Config{
					MessageHandler: func(ctx context.Context, message Message) error {
						return nil
					},
				},
				loggerFactory: func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger) {
					t.Helper()
					log := NewMockConsumerLogger(ctrl)
					return log, log, log
				},
			},
			args: args{
				message: &Message{},
			},
			wantErr: false,
		},
		{
			name: "Good with pausable handler",
			fields: fields{
				config: &Config{
					PausableMessageHandler: func(ctx context.Context, message Message, pauseResumer PauseResumer) error {
						return nil
					},
				},
				loggerFactory: func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger) {
					t.Helper()
					log := NewMockConsumerLogger(ctrl)
					return log, log, log
				},
			},
			args: args{
				message: &Message{},
			},
			wantErr: false,
		},
		{
			name: "With error",
			fields: fields{
				config: &Config{
					MessageHandler: func(ctx context.Context, message Message) error {
						return errors.New("oops")
					},
					Timeout:              2 * time.Second,
					ErrorHandlingTimeout: 2 * time.Second,
				},
				loggerFactory: func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger) {
					t.Helper()
					log := NewMockConsumerLogger(ctrl)
					return log, log, log
				},
			},
			args: args{
				message: &Message{},
			},
			wantErr: true,
		},
		{
			name: "Handler panics",
			fields: fields{
				config: &Config{
					MessageHandler: func(ctx context.Context, message Message) error {
						panic(errors.New("oops"))
					},
					Timeout:              2 * time.Second,
					ErrorHandlingTimeout: 2 * time.Second,
				},
				loggerFactory: func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger) {
					t.Helper()
					log := NewMockConsumerLogger(ctrl)
					log.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)
					return log, log, log
				},
			},
			args: args{
				message: &Message{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, _, de := tt.fields.loggerFactory(t, ctrl)
			c := &KafkaConsumer{
				config:   tt.fields.config,
				debugLog: de,
			}
			err := c.invokeMessageHandler(context.TODO(), tt.args.message)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_invokeNotificationHandler(t *testing.T) {
	type fields struct {
		config        *Config
		loggerFactory func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger)
	}
	type args struct {
		notification string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Good",
			fields: fields{
				config: &Config{
					NotificationHandler: func(message string) {},
				},
				loggerFactory: func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger) {
					t.Helper()
					log := NewMockConsumerLogger(ctrl)
					return log, log, log
				},
			},
			args: args{
				notification: "a",
			},
		},
		{
			name: "Handler panics",
			fields: fields{
				config: &Config{
					NotificationHandler: func(message string) {
						panic(errors.New("oops"))
					},
				},
				loggerFactory: func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger) {
					t.Helper()
					log := NewMockConsumerLogger(ctrl)
					log.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)
					return log, log, log
				},
			},
			args: args{
				notification: "a",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			_, _, de := tt.fields.loggerFactory(t, ctrl)
			c := &KafkaConsumer{
				config:   tt.fields.config,
				debugLog: de,
			}

			c.invokeNotificationHandler(tt.args.notification)
		})
	}
}

func Test_pullOrdered_close(t *testing.T) {
	type fields struct {
		partitionConsumersFactory func(t *testing.T) map[partition]partitionConsumer
	}
	type args struct {
		wait bool
	}
	tests := []struct {
		name                   string
		fields                 fields
		args                   args
		wantPartitionConsumers map[partition]int
	}{
		{
			name: "No wait",
			fields: fields{
				partitionConsumersFactory: func(t *testing.T) map[partition]partitionConsumer {
					c := map[partition]partitionConsumer{
						{
							topic:     "topic",
							partition: 0,
						}: {requests: make(chan *Message, 1), processingDoneWg: &sync.WaitGroup{}}, // simulate a waiting producer with 1 message
						{
							topic:     "topic",
							partition: 1,
						}: {requests: make(chan *Message, 1), processingDoneWg: &sync.WaitGroup{}}, // simulate a waiting producer with 1 message
					}
					for _, messages := range c {
						messages.requests <- &Message{}
					}
					return c
				},
			},
			args: args{
				wait: false,
			},
			wantPartitionConsumers: map[partition]int{},
		},
		{
			name: "Wait",
			fields: fields{
				partitionConsumersFactory: func(t *testing.T) map[partition]partitionConsumer {
					c := map[partition]partitionConsumer{
						{
							topic:     "topic",
							partition: 0,
						}: {requests: make(chan *Message, 1), processingDoneWg: &sync.WaitGroup{}}, // simulate a waiting producer with 1 message
						{
							topic:     "topic",
							partition: 1,
						}: {requests: make(chan *Message, 1), processingDoneWg: &sync.WaitGroup{}}, // simulate a waiting producer with 1 message
					}
					for _, messages := range c {
						messages.requests <- &Message{}
					}
					return c
				},
			},
			args: args{
				wait: true,
			},
			wantPartitionConsumers: map[partition]int{
				{
					topic:     "topic",
					partition: 0,
				}: 1,
				{
					topic:     "topic",
					partition: 1,
				}: 1,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			po := &pullOrdered{
				partMx:             &sync.RWMutex{},
				partitionConsumers: tt.fields.partitionConsumersFactory(t),
			}

			consClone := make(map[partition]partitionConsumer)
			for k, v := range po.partitionConsumers { // copy map as all elements in the original map will be removed
				consClone[k] = v
			}

			po.close(tt.args.wait)

			require.Empty(t, po.partitionConsumers)

			if len(tt.wantPartitionConsumers) > 0 {
				for p, cl := range tt.wantPartitionConsumers {
					cons, ok := consClone[p]
					require.True(t, ok)
					require.Equal(t, cl, len(cons.requests))
				}
			}
		})
	}
}

func Test_pullOrdered_handleAssignedPartitions(t *testing.T) {
	topic := "topic"
	partitionConsumers := map[partition]partitionConsumer{
		{
			topic:     topic,
			partition: 2,
		}: {requests: make(chan *Message), processingDoneWg: &sync.WaitGroup{}},
	}

	parts := kafka.AssignedPartitions{
		Partitions: []kafka.TopicPartition{
			{
				Topic:     &topic,
				Partition: 0,
			},
		},
	}
	processingWG := sync.WaitGroup{}
	msgProc := func(*Message) {
		processingWG.Done()
	}

	t.Run("good", func(t *testing.T) {
		po := &pullOrdered{
			partMx:             &sync.RWMutex{},
			partitionConsumers: partitionConsumers,
		}

		po.handleAssignedPartitions(parts, msgProc)

		require.Len(t, po.partitionConsumers, 1)
		c, ok := po.partitionConsumers[partition{
			topic:     topic,
			partition: 0,
		}]
		require.True(t, ok)

		processingWG.Add(1)
		c.requests <- &Message{}
		waitOrFail(t, &processingWG, 5*time.Second)
		po.close(false)
	})
}

func Test_pullOrdered_handleMessage(t *testing.T) {
	topic := "topic"
	t.Run("good", func(t *testing.T) {
		jobChan := make(chan *Message, 1)
		jobWG := sync.WaitGroup{}
		partitionConsumers := map[partition]partitionConsumer{
			{
				topic:     topic,
				partition: 0,
			}: {requests: jobChan, processingDoneWg: &sync.WaitGroup{}},
		}
		go func() {
			for range jobChan {
				jobWG.Done()
			}
		}()

		po := &pullOrdered{
			partMx:             &sync.RWMutex{},
			partitionConsumers: partitionConsumers,
		}
		jobWG.Add(1)
		err := po.handleMessage(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: 0,
			},
		})
		require.NoError(t, err)
		waitOrFail(t, &jobWG, 5*time.Second)
	})

	t.Run("Bad partition", func(t *testing.T) {
		po := &pullOrdered{
			partMx:             &sync.RWMutex{},
			partitionConsumers: map[partition]partitionConsumer{},
		}

		err := po.handleMessage(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: 0,
			},
		})
		require.Error(t, err)
	})
}

func Test_pullOrdered_handleRevokedPartitions(t *testing.T) {
	topic := "topic"
	t.Run("good", func(t *testing.T) {
		jobChan := make(chan *Message, 1)
		partitionConsumers := map[partition]partitionConsumer{
			{
				topic:     topic,
				partition: 0,
			}: {requests: jobChan, processingDoneWg: &sync.WaitGroup{}},
		}

		po := &pullOrdered{
			partMx:             &sync.RWMutex{},
			partitionConsumers: partitionConsumers,
		}

		po.handleRevokedPartitions(kafka.RevokedPartitions{})

		require.Len(t, po.partitionConsumers, 0)

		_, ok := <-jobChan
		require.False(t, ok)
	})
}

func Test_pullUnOrdered_close(t *testing.T) {
	t.Run("good no wait", func(t *testing.T) {
		q := make(chan *Message, 1)
		pu := &pullUnOrdered{
			workerPool: workerpool.New(1),
			queueMx:    &sync.RWMutex{},
			queue:      q,
		}
		pu.queue <- &Message{}

		pu.close(false)

		require.True(t, pu.workerPool.Stopped())
		require.Nil(t, pu.queue)
		_, ok := <-q
		require.False(t, ok)
	})

	t.Run("good wait", func(t *testing.T) {
		q := make(chan *Message, 1)
		pu := &pullUnOrdered{
			workerPool: workerpool.New(1),
			queueMx:    &sync.RWMutex{},
			queue:      q,
		}
		pu.queue <- &Message{}

		pu.close(true)

		require.True(t, pu.workerPool.Stopped())
		require.Nil(t, pu.queue)
		_, ok := <-q
		require.True(t, ok)
		_, ok = <-q
		require.False(t, ok)
	})
}

func Test_pullUnOrdered_handleAssignedPartitions(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		processingWG := sync.WaitGroup{}
		msgProc := func(*Message) {
			processingWG.Done()
		}
		q := make(chan *Message, 1)
		pu := &pullUnOrdered{
			queueSize:  1,
			workerPool: workerpool.New(1),
			queueMx:    &sync.RWMutex{},
			queue:      q,
		}

		pu.handleAssignedPartitions(kafka.AssignedPartitions{}, msgProc)

		_, ok := <-q
		require.False(t, ok)

		processingWG.Add(1)
		pu.queue <- &Message{}
		waitOrFail(t, &processingWG, 5*time.Second)

		pu.close(false)
	})
}

func Test_pullUnOrdered_handleMessage(t *testing.T) {
	topic := "topic"
	t.Run("good", func(t *testing.T) {
		q := make(chan *Message, 1)
		pu := &pullUnOrdered{
			queueSize:  1,
			workerPool: workerpool.New(1),
			queueMx:    &sync.RWMutex{},
			queue:      q,
		}

		err := pu.handleMessage(&kafka.Message{TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: 0,
		}})

		require.NoError(t, err)
		m, ok := <-pu.queue
		require.True(t, ok)
		require.NotNil(t, m)

		pu.close(false)
	})
	t.Run("no queue", func(t *testing.T) {
		pu := &pullUnOrdered{
			queueSize:  1,
			workerPool: workerpool.New(1),
			queueMx:    &sync.RWMutex{},
			queue:      nil,
		}

		err := pu.handleMessage(&kafka.Message{TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: 0,
		}})

		require.Error(t, err)

		pu.close(false)
	})
}

func Test_pullUnOrdered_handleRevokedPartitions(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		q := make(chan *Message, 1)
		pu := &pullUnOrdered{
			queueSize:  1,
			workerPool: workerpool.New(1),
			queueMx:    &sync.RWMutex{},
			queue:      q,
		}

		pu.handleRevokedPartitions(kafka.RevokedPartitions{})

		require.Nil(t, pu.queue)
		_, ok := <-q
		require.False(t, ok)
	})
}

func waitOrFail(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	testDoneCH := make(chan struct{})
	go func() {
		wg.Wait()
		testDoneCH <- struct{}{}
		close(testDoneCH)
	}()
	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
		t.Fail()
	case <-testDoneCH:
	}
}
