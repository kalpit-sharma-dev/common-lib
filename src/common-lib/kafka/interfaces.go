// Package kafka implements kafka client configuration details
//
// Deprecated: kafka is old implementation of kafka connectivity and should not be used
// except for compatibility with legacy systems.
//
// Use https://gitlab.kksharmadevdev.com/platform/Platform-Infrastructure-lib/tree/master/messaging for all kafka connectivity
// This package is frozen and no new functionality will be added.
package kafka

import (
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/kafka/encode"
)

//go:generate mockgen -package mock -destination=mock/mocks.go . ProducerFactory,ConsumerFactory,ProducerService,ConsumerService,Limiter

// ProducerCommandFactory interface that has base producer command Service
type ProducerCommandFactory interface {
	GetProducerCommandService() ProducerCommand
	GetConfluentProducerCommandService() ProducerCommand
}

// ConsumerCommandFactory interface that has base producer command
type ConsumerCommandFactory interface {
	GetConsumerCommandService() ConsumerCommand
}

// Commmand base interface with basic functions for kafka
type Commmand interface {
	Close() error
}

// ProducerCommand interface that has base producer command
type ProducerCommand interface {
	Commmand
	PushMessage(topicName string, message string) (partition int32, offset int64, err error)
	PushMessageEncoder(topicName string, message encode.Encoder) (partition int32, offset int64, err error)
	NewProducer([]string) error
	IsConnected() bool
}

// Limiter mechanism to reduce the speed of consuming
type Limiter interface {
	IsConsumingAllowed() bool
	Wait()
}

// ConsumerCommand interface that has base consumer command
type ConsumerCommand interface {
	Commmand
	PullMessage(ConsumerHandler)
	LimitedPullMessageNoOffset(ConsumerHandler, Limiter)
	MarkOffset(string, int32, int64)
	NewConsumer([]string, string, []string) error
	NewCustomConsumer(*ConsumerKafkaInOutParams, []string, string, []string) error
	IsConnected() bool
}

// ConsumerCommandSequential interface that adds a sequential consumer command
type ConsumerCommandSequential interface {
	ConsumerCommand
	PullMessageProcessSequentially(ConsumerHandler)
	NewConsumerSafe([]string, string, []string) error
}

// ProducerFactory returns a ProducerImp
type ProducerFactory interface {
	GetProducerService(ProducerConfig) (ProducerService, error)
	GetConfluentProducerService(ProducerConfig) (ProducerService, error)
}

// ConsumerFactory returns a ConsumerImp
type ConsumerFactory interface {
	GetConsumerService(ConsumerConfig) (ConsumerService, error)
}

// Client interface contains commmon methods for kafka clients
type client interface {
	//IsConnected() (result bool)
	CloseConnection() (err error)
}

// ProducerService interface to be implemented by every kafka producer structure
type ProducerService interface {
	client
	//Connect() (err error)
	Push(topicname string, message string) (err error)
	PushEncoder(topicName string, message encode.Encoder) (err error)
}

// ConsumerService interface to be implemented by every kafka consumer structure
type ConsumerService interface {
	client
	//Connect(topics []string) (err error)
	//Pull() (message string, err error)
	PullHandler(consumerHandler ConsumerHandler) error
	PullHandlerWithLimiter(consumerHandler ConsumerHandler, limiter Limiter) error
	Connect(config *ConsumerKafkaInOutParams) error
	MarkOffset(topic string, partition int32, offset int64)
}

// ConsumerServiceWithOrder interface to be implemented by every kafka consumer need strict consume order
type ConsumerServiceWithOrder interface {
	ConsumerService
	PullHandlerSequential(consumerHandler ConsumerHandler) error
}

// ConsumerHandler a message Handler after pulling data from kafka
type ConsumerHandler func(ConsumerMessage)

// ConsumerMessage represents consumed Message
type ConsumerMessage struct {
	Message             string
	Offset              int64
	Partition           int32
	Topic               string
	ReceivedDateTimeUTC time.Time
}

// Error Codes
const (
	ErrorClientNotConnected       = "ErrKafkaClientNotConnected"
	ErrorClientConnected          = "ErrKafkaClientAlreadyConnected"
	ErrorBrokerAddressNotProvided = "ErrKafkaBrokerAddressNotProvided"
	ErrorClientGroupIDNotProvided = "ErrKafkaClientGroupIDNotProvided"
	ErrorTopicsNotProvided        = "ErrKafkaTopicsNotProvided"
)

// ConsumerKafkaInOutParams Config for kafka to be passed from the client
type ConsumerKafkaInOutParams struct {
	ReturnErrors        bool
	ReturnNotifications bool
	OffsetsInitial      int64
	Retention           time.Duration
	Errors              chan error  //to show errors from Kafka
	Notifications       chan string //to show notifications from Kafka
}
