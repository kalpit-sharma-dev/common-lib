package producer

import (
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

// Logger : Logger instance used for logging
// Defaults to Discard
var Logger = logger.DiscardLogger

// Config is a struct used by Kafka producer
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

// NewConfig - returns a configration object having default values
func NewConfig() *Config {
	return &Config{
		TimeoutInSecond:           5,
		CompressionType:           CompressionNone,
		MaxMessageBytes:           1000000,
		ReconnectBackoffMs:        100,
		ReconnectBackoffMaxMs:     10000,
		MessageTimeoutMs:          300000,
		MessageSendMaxRetries:     2,
		RetryBackoffMs:            100,
		QueueBufferingMaxMessages: 100000,
		QueueBufferingMaxKbytes:   1048576,
		QueueBufferingMaxMs:       5,
		EnableIdempotence:         false,
		ProduceChannelSize:        1000000,
	}
}

const (
	// CompressionNone is used for disabling all the Compression
	CompressionNone = "none"
	// CompressionGZIP is used for GZIP Compression
	CompressionGZIP = "gzip"
	// CompressionSnappy is used for Snappy Compression
	CompressionSnappy = "snappy"
	// CompressionLZ4 is used for LZ4 Compression
	CompressionLZ4 = "lz4"
)
