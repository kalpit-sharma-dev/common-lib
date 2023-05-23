package consumer

import (
	"context"
	"time"
)

type commitMode int

type consumerMode int

// Config is a struct used by Kafka Consumer
type Config struct {
	// Address: Kafka Broker IPs
	Address []string

	// Group: A Consumer Group name
	Group string

	// Topics: Topics to be consumed
	Topics []string

	// SubscriberPerCore: Number of maximum go routines for processing per Core
	SubscriberPerCore int

	// CommitMode: Specify which commit mode you want to use
	// OnPull means offset will be committed on Pull before message processing starts
	// OnMessageCompletion means offset will be committed on after message processing completes
	// Default: OnPull
	CommitMode commitMode

	// ConsumerMode: Specify which consumer mode you want to use
	// PullUnOrdered : Messages will be processed in batches
	// PullOrdered : This Mode claims that message will not lost under any system failure.
	//
	// NOTE: This needs more partition as messages are processed sequentially / partition.
	// Default: PullUnOrdered
	ConsumerMode consumerMode

	// Timeout: After this duration message offset will be committed without releasing worker
	Timeout time.Duration

	// ErrorHandlingTimeout: Amount of time to wait for the error handler to return.
	// Default: 1m.
	ErrorHandlingTimeout time.Duration

	// MessageHandler: This will get used as a callback function for message handling (must be thread safe)
	// Incompatible with PausableMessageHandler. There must be either MessageHandler or PausableMessageHandler provided to the Config. NOT both!
	MessageHandler func(context.Context, Message) error

	// PausableMessageHandler: This will get used as a callback function for message handling (must be thread safe) with access to the Pause/Resume functionality.
	// Incompatible with MessageHandler. There must be either MessageHandler or PausableMessageHandler provided to the Config. NOT both!
	PausableMessageHandler func(context.Context, Message, PauseResumer) error

	// ErrorHandler: If available, any errors that occurred while consuming are returned on this handler (must be thread safe)
	ErrorHandler func(context.Context, error, *Message)

	// NotificationHandler: If available, rebalance notification will be returned on this handler
	NotificationHandler func(string)

	// OffsetsInitial: The initial offset to use if no offset was previously committed.
	// Should be OffsetNewest or OffsetOldest.
	// Default: OffsetNewest.
	OffsetsInitial int64

	// RetryDelay: This flag is used in case you want to wait for some time before retrying the the failed message.
	// Default: 30s
	RetryDelay time.Duration

	// RetryCount: Specify how meany times a message should be retried before dropping the message
	// Default: 10
	RetryCount int64

	// TransactionID: Default transaction id to be used by the consumer
	TransactionID string

	// Partitions: Max Number of partitions to be consumed by consumer
	// NOTE: This number should always be bigger than the number of partitions available in the production for all the topics
	// Default: 500.
	Partitions int

	// MaxQueueSize: Max messages to keep in local queue
	// NOTE: the greater this is, the more memory the service uses processing messages does not keep up
	// Default: 100
	MaxQueueSize int

	// EmptyQueueWaitTime: specifies the time to wait when we encounter an empty queue of messages
	// NOTE: this is to keep CPU from spiking at the cost of locking the main process thread, keep it low
	// Default: 500ms
	EmptyQueueWaitTime time.Duration

	// CommitIntervalMs: interval between committing stored offsets
	// Default: 5000
	CommitIntervalMs int

	// EnableKafkaLogs: whether to enable logging of kafka internals.
	// Default: false
	EnableKafkaLogs bool

	// LogLevel: Kafka log level if enabled. 0..7.
	// Default: 3
	LogLevel int

	// LoggedComponents: comma separated list of components to be logged.
	// Allowed: generic, broker, topic, metadata, feature, queue, msg, protocol, cgrp, security, fetch, interceptor, plugin, consumer, admin, eos, mock, assignor, conf, all.
	// Detailed Producer "debugging: broker,topic,msg". Consumer: "consumer,cgrp,topic,fetch"
	// Default: all
	LoggedComponents string

	// ErrorLog is used to log errors.
	ErrorLog ConsumerLogger
	// InfoLog is used to log info logs.
	InfoLog ConsumerLogger
	// DebugLog is used to log debug logs.
	DebugLog ConsumerLogger
}

// NewConfig is a function to return default consumer configuration
func NewConfig() *Config {
	return &Config{
		SubscriberPerCore:    20,
		CommitMode:           OnPull,
		ConsumerMode:         PullUnOrdered,
		OffsetsInitial:       OffsetNewest,
		Timeout:              time.Minute,
		ErrorHandlingTimeout: time.Minute,
		RetryCount:           10,
		RetryDelay:           30 * time.Second,
		Partitions:           500,
		MaxQueueSize:         100,
		EmptyQueueWaitTime:   500 * time.Millisecond,
		CommitIntervalMs:     5000,
		EnableKafkaLogs:      false,
		LogLevel:             3,
		LoggedComponents:     "all",
	}
}
