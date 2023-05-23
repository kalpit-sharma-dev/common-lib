package consumer

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestManualKafkaConsumer_Close(t *testing.T) {
	type fields struct {
		consumer           func(t *testing.T, ctrl *gomock.Controller) consumer
		pausedPartitionsMx *sync.Mutex
		pausedPartitions   map[partition]kafka.TopicPartition
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Good",
			fields: fields{
				consumer: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Close().Return(nil).Times(1)
					return c
				},
				pausedPartitionsMx: &sync.Mutex{},
				pausedPartitions: map[partition]kafka.TopicPartition{
					{"topic1", 0}: {},
					{"topic2", 0}: {},
				},
			},
			wantErr: false,
		},
		{
			name: "Kafka return error",
			fields: fields{
				consumer: func(t *testing.T, ctrl *gomock.Controller) consumer {
					t.Helper()
					c := NewMockconsumer(ctrl)
					c.EXPECT().Close().Return(errors.New("oops")).Times(1)
					return c
				},
				pausedPartitionsMx: &sync.Mutex{},
				pausedPartitions: map[partition]kafka.TopicPartition{
					{"topic1", 0}: {},
					{"topic2", 0}: {},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			c := &ManualKafkaConsumer{
				consumer:           tt.fields.consumer(t, ctrl),
				pausedPartitionsMx: tt.fields.pausedPartitionsMx,
				pausedPartitions:   tt.fields.pausedPartitions,
			}

			err := c.Close()

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Empty(t, c.pausedPartitions)
		})
	}
}

func TestManualKafkaConsumer_CommitOffset(t *testing.T) {
	topic := "topic"
	type fields struct {
		consumer consumer
	}
	type args struct {
		topic     string
		partition int32
		offset    int64
	}
	tests := []struct {
		name          string
		fieldsFactory func(t *testing.T, ctrl *gomock.Controller) fields
		args          args
		wantErr       bool
	}{
		{
			name: "Good",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()
				c := NewMockconsumer(ctrl)
				c.EXPECT().CommitOffsets([]kafka.TopicPartition{{
					Topic:     &topic,
					Partition: 0,
					Offset:    kafka.Offset(1),
				}}).Return(nil, nil).Times(1)

				return fields{
					consumer: c,
				}
			},
			args: args{
				topic:     topic,
				partition: 0,
				offset:    0,
			},
			wantErr: false,
		},
		{
			name: "Kafka returns error",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				c := NewMockconsumer(ctrl)
				c.EXPECT().CommitOffsets([]kafka.TopicPartition{{
					Topic:     &topic,
					Partition: 0,
					Offset:    kafka.Offset(1),
				}}).Return(nil, errors.New("oops")).Times(1)

				return fields{
					consumer: c,
				}
			},
			args: args{
				topic:     topic,
				partition: 0,
				offset:    0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			f := tt.fieldsFactory(t, ctrl)

			c := &ManualKafkaConsumer{
				consumer: f.consumer,
			}

			gotErr := c.CommitOffset(tt.args.topic, tt.args.partition, tt.args.offset)
			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}
		})
	}
}

func TestManualKafkaConsumer_Health(t *testing.T) {
	health := Health{
		ConnectionState: false,
		Address:         []string{"127.0.0.1"},
		Topics:          []string{"aa"},
		Group:           "group",
	}
	t.Run("Get Health", func(t *testing.T) {
		c := &ManualKafkaConsumer{
			healthMx: &sync.RWMutex{},
			health:   &health,
		}
		got, err := c.Health()

		require.NoError(t, err)
		require.Equal(t, health, got)
	})
}

func TestManualKafkaConsumer_Pause(t *testing.T) {
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

			c := &ManualKafkaConsumer{
				consumer: tt.fields.consumerFactory(t, ctrl),

				pausedPartitionsMx: &sync.Mutex{},
				pausedPartitions:   tt.fields.pausedPartitions,
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

func TestManualKafkaConsumer_PauseAll(t *testing.T) {
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

			c := &ManualKafkaConsumer{
				consumer:           tt.fields.consumerFactory(t, ctrl),
				pausedPartitionsMx: &sync.Mutex{},
				pausedPartitions:   tt.fields.pausedPartitions,
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

func TestManualKafkaConsumer_Poll(t *testing.T) {
	type fields struct {
		consumer consumer
		debugLog ConsumerLogger
	}
	tests := []struct {
		name                      string
		fieldsFactory             func(t *testing.T, ctrl *gomock.Controller) fields
		want                      *kafka.Message
		wantHealthConnectionState bool
	}{
		{
			name: "Good",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				msg := kafka.Message{
					Key: []byte("key"),
				}
				c := NewMockconsumer(ctrl)
				c.EXPECT().Poll(gomock.Any()).Return(&msg).Times(1)

				return fields{
					consumer: c,
				}
			},
			want: &kafka.Message{
				Key: []byte("key"),
			},
			wantHealthConnectionState: true,
		},
		{
			name: "Unsupported message type - return nil",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				c := NewMockconsumer(ctrl)
				c.EXPECT().Poll(gomock.Any()).Return(kafka.AssignedPartitions{}).Times(1)

				return fields{
					consumer: c,
				}
			},
			want:                      nil,
			wantHealthConnectionState: true,
		},
		{
			name: "Non-ErrAllBrokersDown error - return nil",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				msg := kafka.NewError(kafka.ErrBadMsg, "err", false)
				c := NewMockconsumer(ctrl)
				c.EXPECT().Poll(gomock.Any()).Return(msg).Times(1)

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

				return fields{
					consumer: c,
					debugLog: l,
				}
			},
			want:                      nil,
			wantHealthConnectionState: true,
		},
		{
			name: "ErrAllBrokersDown error - set ConnectionState to false & return nil",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				msg := kafka.NewError(kafka.ErrAllBrokersDown, "err", false)
				c := NewMockconsumer(ctrl)
				c.EXPECT().Poll(gomock.Any()).Return(msg).Times(1)

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

				return fields{
					consumer: c,
					debugLog: l,
				}
			},
			want:                      nil,
			wantHealthConnectionState: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			f := tt.fieldsFactory(t, ctrl)
			c := &ManualKafkaConsumer{
				config:   &Config{},
				consumer: f.consumer,
				healthMx: &sync.RWMutex{},
				health:   newHealth(),
				debugLog: f.debugLog,
			}
			c.healthMx.Lock()
			c.health.ConnectionState = true
			c.healthMx.Unlock()

			got := c.Poll(100)

			require.Equal(t, tt.want, got)
			c.healthMx.RLock()
			defer c.healthMx.RUnlock()
			require.Equal(t, tt.wantHealthConnectionState, c.health.ConnectionState)
		})
	}
}

func TestManualKafkaConsumer_Resume(t *testing.T) {
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

			c := &ManualKafkaConsumer{
				consumer: tt.fields.consumerFactory(t, ctrl),

				pausedPartitionsMx: &sync.Mutex{},
				pausedPartitions:   tt.fields.pausedPartitions,
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

func TestManualKafkaConsumer_ResumeAll(t *testing.T) {
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

			c := &ManualKafkaConsumer{
				consumer:           tt.fields.consumerFactory(t, ctrl),
				pausedPartitionsMx: &sync.Mutex{},
				pausedPartitions:   tt.fields.pausedPartitions,
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

func TestManualKafkaConsumer_debugf(t *testing.T) {
	t.Run("debugf", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		l := NewMockConsumerLogger(ctrl)
		l.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

		c := &ManualKafkaConsumer{
			debugLog: l,
		}

		c.debugf("aaa %s", "bbb")
	})
}

func TestManualKafkaConsumer_errorf(t *testing.T) {
	t.Run("errorf", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		l := NewMockConsumerLogger(ctrl)
		l.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

		c := &ManualKafkaConsumer{
			errorLog: l,
		}

		c.errorf("aaa %s", "bbb")
	})
}

func TestManualKafkaConsumer_infof(t *testing.T) {
	t.Run("infof", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		l := NewMockConsumerLogger(ctrl)
		l.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

		c := &ManualKafkaConsumer{
			infoLog: l,
		}

		c.infof("aaa %s", "bbb")
	})
}

func TestManualKafkaConsumer_invokeErrorHandler(t *testing.T) {
	type fields struct {
		config        *Config
		loggerFactory func(t *testing.T, ctrl *gomock.Controller) (ConsumerLogger, ConsumerLogger, ConsumerLogger)
	}
	type args struct {
		err error
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
				err: errors.New("oops"),
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
				err: errors.New("oops"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			er, _, de := tt.fields.loggerFactory(t, ctrl)
			c := &ManualKafkaConsumer{
				config:   tt.fields.config,
				errorLog: er,
				debugLog: de,
			}

			c.invokeErrorHandler(tt.args.err)
		})
	}
}

func TestManualKafkaConsumer_invokeNotificationHandler(t *testing.T) {
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

			_, _, de := tt.fields.loggerFactory(t, ctrl)
			c := &ManualKafkaConsumer{
				config:   tt.fields.config,
				debugLog: de,
			}

			c.invokeNotificationHandler(tt.args.notification)
		})
	}
}

func TestManualKafkaConsumer_logKafkaLogs(t *testing.T) {
	type fields struct {
		errorLog ConsumerLogger
		infoLog  ConsumerLogger
		debugLog ConsumerLogger
	}
	type args struct {
		logEv kafka.LogEvent
	}
	tests := []struct {
		name          string
		fieldsFactory func(t *testing.T, ctrl *gomock.Controller) fields
		args          args
	}{
		{
			name: "Log level 0",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

				return fields{
					errorLog: l,
				}
			},
			args: args{
				logEv: kafka.LogEvent{
					Name:    "0",
					Tag:     "0",
					Message: "0",
					Level:   0,
				},
			},
		},
		{
			name: "Log level 1",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

				return fields{
					errorLog: l,
				}
			},
			args: args{
				logEv: kafka.LogEvent{
					Name:    "1",
					Tag:     "1",
					Message: "1",
					Level:   1,
				},
			},
		},
		{
			name: "Log level 2",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

				return fields{
					errorLog: l,
				}
			},
			args: args{
				logEv: kafka.LogEvent{
					Name:    "2",
					Tag:     "2",
					Message: "2",
					Level:   2,
				},
			},
		},
		{
			name: "Log level 3",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

				return fields{
					infoLog: l,
				}
			},
			args: args{
				logEv: kafka.LogEvent{
					Name:    "3",
					Tag:     "3",
					Message: "3",
					Level:   3,
				},
			},
		},
		{
			name: "Log level 4",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

				return fields{
					infoLog: l,
				}
			},
			args: args{
				logEv: kafka.LogEvent{
					Name:    "4",
					Tag:     "4",
					Message: "4",
					Level:   4,
				},
			},
		},
		{
			name: "Log level 5",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

				return fields{
					infoLog: l,
				}
			},
			args: args{
				logEv: kafka.LogEvent{
					Name:    "5",
					Tag:     "5",
					Message: "5",
					Level:   5,
				},
			},
		},
		{
			name: "Log level 6",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

				return fields{
					infoLog: l,
				}
			},
			args: args{
				logEv: kafka.LogEvent{
					Name:    "6",
					Tag:     "6",
					Message: "6",
					Level:   6,
				},
			},
		},
		{
			name: "Log level 7",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

				return fields{
					debugLog: l,
				}
			},
			args: args{
				logEv: kafka.LogEvent{
					Name:    "7",
					Tag:     "7",
					Message: "7",
					Level:   7,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			f := tt.fieldsFactory(t, ctrl)

			c := &ManualKafkaConsumer{
				errorLog: f.errorLog,
				infoLog:  f.infoLog,
				debugLog: f.debugLog,
			}

			lc := make(chan kafka.LogEvent, 1)
			lc <- tt.args.logEv
			close(lc)

			c.logKafkaLogs(lc)
		})
	}
}

func TestManualKafkaConsumer_processAssignedPartitions(t *testing.T) {
	topic := "topic"
	assignedPartitions := kafka.AssignedPartitions{
		Partitions: []kafka.TopicPartition{
			{
				Topic:     &topic,
				Partition: 0,
				Offset:    1,
			},
		},
	}

	type fields struct {
		debugLog ConsumerLogger
		consumer consumer
	}
	tests := []struct {
		name                     string
		fieldsFactory            func(t *testing.T, ctrl *gomock.Controller) fields
		wantHeathConnectionState bool
	}{
		{
			name: "Good",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				c := NewMockconsumer(ctrl)
				c.EXPECT().Resume(gomock.Any()).Return(nil).Times(1)
				c.EXPECT().Assign(assignedPartitions.Partitions).Return(nil).Times(1)

				return fields{
					consumer: c,
				}
			},
			wantHeathConnectionState: true,
		},
		{
			name: "Resume errors",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				c := NewMockconsumer(ctrl)
				c.EXPECT().Resume(gomock.Any()).Return(errors.New("oops")).Times(1)
				c.EXPECT().Assign(assignedPartitions.Partitions).Return(nil).Times(1)

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

				return fields{
					consumer: c,
					debugLog: l,
				}
			},
			wantHeathConnectionState: true,
		},
		{
			name: "Assign errors",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				c := NewMockconsumer(ctrl)
				c.EXPECT().Resume(gomock.Any()).Return(nil).Times(1)
				c.EXPECT().Assign(assignedPartitions.Partitions).Return(errors.New("oops")).Times(1)

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

				return fields{
					consumer: c,
					debugLog: l,
				}
			},
			wantHeathConnectionState: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			f := tt.fieldsFactory(t, ctrl)

			c := &ManualKafkaConsumer{
				config:             &Config{},
				consumer:           f.consumer,
				debugLog:           f.debugLog,
				healthMx:           &sync.RWMutex{},
				health:             newHealth(),
				pausedPartitionsMx: &sync.Mutex{},
				pausedPartitions:   map[partition]kafka.TopicPartition{{"topic", 0}: {}},
			}

			c.processAssignedPartitions(assignedPartitions)
			c.healthMx.RLock()
			defer c.healthMx.RUnlock()
			require.Equal(t, tt.wantHeathConnectionState, c.health.ConnectionState)
		})
	}
}

func TestManualKafkaConsumer_processRevokedPartitions(t *testing.T) {
	type fields struct {
		debugLog ConsumerLogger
		consumer consumer
	}
	tests := []struct {
		name          string
		fieldsFactory func(t *testing.T, ctrl *gomock.Controller) fields
	}{
		{
			name: "Good",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				c := NewMockconsumer(ctrl)
				c.EXPECT().Resume(gomock.Any()).Return(nil).Times(1)
				c.EXPECT().Unassign().Return(nil).Times(1)

				return fields{
					consumer: c,
				}
			},
		},
		{
			name: "Resume errors",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				c := NewMockconsumer(ctrl)
				c.EXPECT().Resume(gomock.Any()).Return(errors.New("oops")).Times(1)
				c.EXPECT().Unassign().Return(nil).Times(1)

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

				return fields{
					consumer: c,
					debugLog: l,
				}
			},
		},
		{
			name: "Unassign errors",
			fieldsFactory: func(t *testing.T, ctrl *gomock.Controller) fields {
				t.Helper()

				c := NewMockconsumer(ctrl)
				c.EXPECT().Resume(gomock.Any()).Return(nil).Times(1)
				c.EXPECT().Unassign().Return(errors.New("oops")).Times(1)

				l := NewMockConsumerLogger(ctrl)
				l.EXPECT().Printf(gomock.Any(), gomock.Any()).Times(1)

				return fields{
					consumer: c,
					debugLog: l,
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			f := tt.fieldsFactory(t, ctrl)

			c := &ManualKafkaConsumer{
				config:             &Config{},
				consumer:           f.consumer,
				debugLog:           f.debugLog,
				pausedPartitionsMx: &sync.Mutex{},
				pausedPartitions:   map[partition]kafka.TopicPartition{{"topic", 0}: {}},
			}

			c.processRevokedPartitions()
		})
	}
}
