package producer

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

//go:generate mockgen -package mock_producer -source=producer.go -destination=mocks/producer_mock.go .

// Error Codes
var (
	// ErrPublishMessageNotAvailable : Error message not available
	ErrPublishMessageNotAvailable = errors.New("Publish:Message:Not.Available")

	// ErrPublishSendMessageRecovered : Error publish message thrown panic
	ErrPublishSendMessageRecovered = errors.New("Publish:SendMessage:Recovered")

	// ErrPublishSendMessageTimeout : Publish message timed-out
	ErrPublishSendMessageTimeout = errors.New("Publish:SendMessage:Timeout")

	// ErrPublishSendMessageFatal : Publish message is disabled when kafka producer encounters fatal error
	ErrPublishSendMessageFatal = errors.New("Publish:SendMessage:Fatal")
)

// Producer interface
// TODO: rename to SyncProducer?
type Producer interface {
	// Produce: publish kafka messages with context
	ProduceWithReport(ctx context.Context, transaction string, messages ...*Message) ([]*DeliveryReport, error)
	// Produce: publish kafka messages with context
	Produce(ctx context.Context, transaction string, messages ...*Message) error
	// Close
	Close()
	// Health returns producer's health
	Health() (*Health, error)
}

// StatusableProducer is a producer that can return fatal error as status.
type StatusableProducer interface {
	Producer
	GetFatalError() error
}

// AsyncProducer interface
type AsyncProducer interface {
	// ProduceChannel
	ProduceChannel() chan<- *Message
	// EventsChannel
	DeliveryReportChannel() <-chan *DeliveryReport
	// Flush all messages in queue
	Flush(timeoutMs int) int
	// Close
	Close()
	// Health returns producer's health
	Health() (*Health, error)
}

// TopicPartition struct
type TopicPartition struct {
	Topic     string
	Partition int32
	Offset    int64
}

// DeliveryReport represents the result of a produced message
type DeliveryReport struct {
	TopicPartition TopicPartition
	Message        *Message
	Error          error
}

type DeliveryError struct {
	FailedReports []DeliveryReport
}

// Error so that NotFoundError implements error interface.
func (e *DeliveryError) Error() string {
	if e == nil || len(e.FailedReports) == 0 {
		return ""
	}
	errs := make([]string, 0, len(e.FailedReports))
	for i := range e.FailedReports {
		err := e.FailedReports[i].Error
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	return strings.Join(errs, "; ")
}

type produceResult struct {
	deliveryReport []*DeliveryReport
	err            error
}

// Type denotes kafka producer to be used.
// Deprecated: used in a faulty function SyncProducer() which should not be used anymore.
type Type string

const (
	// RegularKafkaProducer denotes normal kafka producer
	// Deprecated: used in a faulty function SyncProducer() which should not be used anymore.
	RegularKafkaProducer Type = "regular"

	// BigKafkaProducer denotes kafka producer capable of publishing big messages
	// Deprecated: used in a faulty function SyncProducer() which should not be used anymore.
	BigKafkaProducer Type = "big"
)

// Deprecated: used in a faulty function SyncProducer() which should not be used anymore.
var (
	kafkaProducer = map[Type]StatusableProducer{}
	producerMutex sync.Mutex
)

// SyncProducer - Return instance of Sync Producer.
// Deprecated: use NewSyncProducer() instead to create new Kafka producer.
var SyncProducer = func(producerType Type, config *Config) (StatusableProducer, error) {
	if kafkaProducer[producerType] == nil {
		producerMutex.Lock()
		defer producerMutex.Unlock()
		if kafkaProducer[producerType] == nil {
			p, err := NewSyncProducer(config)
			if err != nil {
				return nil, err
			}
			kafkaProducer[producerType] = p
		}
	}
	return kafkaProducer[producerType], nil
}

func newKafkaProducer(config *Config, stateChanged func(bool)) (*kafka.Producer, error) {
	logsChan := make(chan kafka.LogEvent, 10000)
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":            strings.Join(config.Address, ","),
		"request.timeout.ms":           config.TimeoutInSecond * 1000,
		"message.max.bytes":            config.MaxMessageBytes,
		"compression.codec":            config.CompressionType,
		"reconnect.backoff.ms":         config.ReconnectBackoffMs,
		"reconnect.backoff.max.ms":     config.ReconnectBackoffMaxMs,
		"message.timeout.ms":           config.MessageTimeoutMs,
		"message.send.max.retries":     config.MessageSendMaxRetries,
		"retry.backoff.ms":             config.RetryBackoffMs,
		"queue.buffering.max.messages": config.QueueBufferingMaxMessages,
		"queue.buffering.max.kbytes":   config.QueueBufferingMaxKbytes,
		"queue.buffering.max.ms":       config.QueueBufferingMaxMs,
		"enable.idempotence":           config.EnableIdempotence,
		"go.produce.channel.size":      config.ProduceChannelSize,
		"go.logs.channel.enable":       true,
		"go.logs.channel":              logsChan,
		"debug":                        "broker",
	})
	if err != nil {
		close(logsChan)
		return nil, err
	}
	// Until we get something better from the lib (currently no reconnect events, just disconnects)
	go func(lc chan kafka.LogEvent, schg func(bool)) {
	ELOOP:
		for {
			select {
			case e, ok := <-logsChan:
				if !ok {
					break ELOOP
				}
				if strings.HasSuffix(e.Message, " -> UP") {
					schg(true)
				} else if strings.HasSuffix(e.Message, " -> DOWN") {
					schg(false)
				}
			}
		}
	}(logsChan, stateChanged)
	return p, nil
}
