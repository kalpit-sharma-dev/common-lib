# Message Producer

Producer is a wrapper on top of [Confluent Kafka client](https://github.com/confluentinc/confluent-kafka-go) 
that provides additional tools for publishing messages such as waiting for delivery
confirmation.

## Additional details

This package provides additional features on top of Confluent:

- Choice between a Synchronous producer (will wait for confirmation it has been
  sent to Kafka) or Asynchronous (fire and forget)
  - Synchronous allows publishing a message using Context to enable timeout
    for a message instead of waiting for longer to avoid an increase in memory consumption
- `Health` - This will provide the Kafka producer connection status.

## Configuration

Configuration used for creation of Kafka producer instance

```go
type Config struct {
	// Address: is a Kafka Broker IPs (bootstrap.servers)
	Address []string

	// TimeoutInSecond: timeout for producer connectivity (request.timeout.ms)
	// Default: 5s
	TimeoutInSecond int

	// MaxMessageBytes : The maximum permitted size of a message. Should be
	// set equal to or smaller than the broker's `message.max.bytes`.
	// Default: 1000000
	MaxMessageBytes int

	// CompressionType: represents the compression codec to use for messages (compression.codec)
	// Default: None
	CompressionType string

	// ReconnectBackoffMs: Ms to exponentially backoff when attempting to reconnect (reconnect.backoff.ms)
	// Default: 100ms
	ReconnectBackoffMs int

	// ReconnectBackoffMaxMs: Max ms to wait between reconnect attempts (reconnect.backoff.max.ms)
	// Default: 10000ms
	ReconnectBackoffMaxMs int

	// MessageTimeoutMs: Local message timeout to limit the wait for successful delivery (message.timeout.ms)
	// Default: 300000ms
	MessageTimeoutMs int

	// MessageSendMaxRetries: The times to retry sending a failing message (message.send.max.retries)
	// Default: 2
	MessageSendMaxRetries int

	// RetryBackoffMs: The backoff time before retrying a protocol request (retry.backoff.ms)
	// Default: 100ms
	RetryBackoffMs int

	// QueueBufferingMaxMessages: Max number of messages in producer queue (queue.buffering.max.messages)
	// Note: This is shared by all topics/partitions
	// Default: 100000
	QueueBufferingMaxMessages int

	// QueueBufferingMaxKbytes: Max total message size sum in producer queue (queue.buffering.max.kbytes)
	// Note: This is higher priority than QueueBufferingMaxMessages
	// Default: 1048576
	QueueBufferingMaxKbytes int

	// QueueBufferingMaxMs: Delay to wait for producer queue before constructing batches (queue.buffering.max.ms)
	// Note: This comes at the expense of delivery latency
	// Default: 5ms
	QueueBufferingMaxMs int

	// EnableIdempotence: When set to `true`, the producer will ensure that messages are successfully produced
	// exactly once and in the original produce order (enable.idempotence)
	// Default: false
	EnableIdempotence bool

	// ProduceChannelSize: Max size of produce channel (go.produce.channel.size)
	// Note: Blocks main thead when reached
	// Default: 1000000
	ProduceChannelSize int
}
```

Creation of the configuration:

```go
producer.NewConfig() *Config
```

## Message

Message is the collection of elements passed to the Producer in order to send a
message to a Kakfa topic. Message also allows us to add header values.

```go
type Message struct {
	// Topic: The Kafka topic for this message
	Topic string
	// Key: The partitioning key for this message
	Key []byte
	// Value: The actual message to store in Kafka
	Value []byte
}
```

The headers are key-value pairs that are transparently passed on the message object:

```go
AddHeader(key string, value string)
```

## Contracts

The syncronous producer (`producer.SyncProducer`) allows us to publish any
message to a given Kafka topic with context:

```go
	// Produce: publish kafka messages with context
	ProduceWithReport(ctx context.Context, transaction string, messages ...*Message) ([]*DeliveryReport, error)
	// Produce: publish kafka messages with context
	Produce(ctx context.Context, transaction string, messages ...*Message) error
```

The asynchronous producer (`producer.NewAsyncProducer`) provides a channel
to send messages to and a channel to access their status for being
delivered:

```go
	// ProduceChannel
	ProduceChannel() chan<- *Message
	// EventsChannel
	DeliveryReportChannel() <-chan *DeliveryReport
```

Both producer types also have the following additional utilities:

```go
	// Close
	Close()
	// Health returns producer's health
	Health() (*Health, error)
}
```

## Reported Errors

Following errors are reported by the Kafka producer while publishing a message
to Kakfa:

```go
	// ErrPublishMessageNotAvailable : Error message not available
	ErrPublishMessageNotAvailable = errors.New("Publish:Message:Not.Available")

	// ErrPublishSendMessageRecovered : Error publish message thrown panic
	ErrPublishSendMessageRecovered = errors.New("Publish:SendMessage:Recovered")

	// ErrPublishSendMessageTimeout : Publish message timed-out
	ErrPublishSendMessageTimeout = errors.New("Publish:SendMessage:Timeout")

	// ErrPublishSendMessageFatal : Publish message is disabled when kafka producer encounters fatal error
	ErrPublishSendMessageFatal = errors.New("Publish:SendMessage:Fatal")
```

## Example

Simple example to create producer instance

```go
cfg := producer.NewConfig()
cfg.Address = ctx.GlobalStringSlice("address")
producer, err := producer.SyncProducer(producer.RegularKafkaProducer, cfg)
```

More complex scenarios can be viewed in [the runnable samples](examples/).
