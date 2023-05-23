# DEPRECATION NOTICE

`messaging-sarama/publisher` is deprecated, consider using `messaging/producer` instead

# Publisher

Publisher is a wrapper on top of [Sarama](https://github.com/Shopify/sarama) → A Kafka client library

- [MIT License](https://github.com/Shopify/sarama/blob/master/LICENSE)

## Additional details on top of Sarama

This package is providing below additional features on top of Sarama

- Reconnect to Kafka
  - Reconnect to another Kafka instance as soon as any one node goes down or stopped responding
- Publish a message using Context
  - Publish a message using to enable timeout for an messaes instead of waiting for longer to avoid increase in memory consumption

## Limitation

- Today only support Sync producer, Async need to implementation in case needed for any MS

## Configuration

Configuration used for creation of Kafka producer instance

```go
type Config struct {
	// Address : is a Kafka Broker IPs
	Address []string

	// TimeoutInSecond - timeout for producer connectivity
	// defaults to 3 Second
	TimeoutInSecond int64

	// MaxMessageBytes : The maximum permitted size of a message (defaults to 1000000). Should be
	// set equal to or smaller than the broker's `message.max.bytes`.
	MaxMessageBytes int

	// CompressionType represents the various compression codecs recognized by Kafka in messages.
	CompressionType CompressionType

	// ReconnectIntervalInSecond - retry interval for Kafka reconnect
	// defaults to TimeoutInSecond * 4
	ReconnectIntervalInSecond int64

	// MaxReconnectRetry - Max number or retry to connect on Kafka
	// defaults to 20
	MaxReconnectRetry int

	// CleanupTimeInSecond - wait time before starting cleanup
	// defaults to TimeoutInSecond * 5
	CleanupTimeInSecond int64

	// CircuitBreaker - Confuguration for the Circuit breaker
	// defaults to - circuit.New()
	CircuitBreaker *circuit.Config

	// CommandName - Name for broker command
	// defaults to - Broker-Command
	CommandName string

	// ValidErrors - List of error to participates in the Circuit state calculation
	// Default values are -
	// ErrClosedClient, ErrOutOfBrokers, ErrNotConnected, ErrShuttingDown
	// ErrControllerNotAvailable, ErrNoTopicsToUpdateMetadata, ErrPublishSendMessageTimeout
	ValidErrors []string
}
```

## Creation of Configuration
```go
publisher.NewConfig() *Config
```
## Message

Message is the collection of elements passed to the Producer in order to send a message to a Kakfa topic

- Message also allows us to add header values

```go
type Message struct {
	// Topic - The Kafka topic for this message.
	Topic string
	// Key - The partitioning key for this message. Pre-existing Encoders include
	// StringEncoder and ByteEncoder.
	Key Encoder
	// Value - The actual message to store in Kafka. Pre-existing Encoders include
	// StringEncoder and ByteEncoder.
	Value Encoder
}
```

AddHeader - The headers are key-value pairs that are transparently passed on Message Object

```go
AddHeader(Key []byte, Value []byte) {
```

## Contracts

Contracts to create Producer instance, Reconnect, and Clean existing connection

```go
// SyncProducer - Return instance of Sync Producer
publisher.SyncProducer(producerType ProducerType, cfg *Config) (Producer, error)

// ReConnect - Function to Reconnect Kafka nodes
publisher.ReConnect(transaction string, cfg *Config)

// Connected - Used for stopping reconnect process as soon as we find MS is connected to Kafka and producting messages
publisher.Connected(transaction string, cfg *Config)

// Clean - Clean all Kafka connection, used in case of graceful shutdown
publisher.Clean(transaction string)
```

Publisher allows us to publish any message to given kafka topic with context

```go
Publish(ctx context.Context, transaction string, messages ...*Message) error
```

## Reported Errors

Following errors are reported by the Kafka producer while publishing a message to Kakfa

```go
// ErrPublishMessageNotAvailable : Error message not available
ErrPublishMessageNotAvailable = errors.New("Publish:Message:Not.Available")

// ErrPublishSendMessageRecovered : Error publish message thrown panic
ErrPublishSendMessageRecovered = errors.New("Publish:SendMessage:Recovered")

// ErrPublishSendMessageTimeout : Publish message timed-out
ErrPublishSendMessageTimeout = errors.New("Publish:SendMessage:Timeout")
```

## Example

Simple example to create producer instance

```go
log := logger.Get()
cfg := publisher.NewConfig()
cfg.Address = ctx.GlobalStringSlice("address")
producer, err := publisher.SyncProducer(publisher.RegularKafkaProducer, cfg)
```

[Complete example](/messaging/publisher/example)
