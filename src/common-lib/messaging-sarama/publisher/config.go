package publisher

import (
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

// Logger : Logger instance used for logging
// Defaults to Discard
var Logger = logger.DiscardLogger

// Config is a struct used by Kafka producer
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

	// CircuitProneErrors - List of error to participates in the Circuit state calculation
	// Default values are -
	// ErrClosedClient, ErrOutOfBrokers, ErrNotConnected, ErrShuttingDown
	// ErrControllerNotAvailable, ErrNoTopicsToUpdateMetadata, ErrPublishSendMessageTimeout
	// You must provide at least one broker address
	CircuitProneErrors []string
}

// NewConfig - returns a configration object having default values
func NewConfig() *Config {
	return &Config{
		TimeoutInSecond: 3, MaxMessageBytes: 1000, ReconnectIntervalInSecond: 12,
		MaxReconnectRetry: 12, CleanupTimeInSecond: 15, CircuitBreaker: &circuit.Config{
			Enabled: true, TimeoutInSecond: 1, MaxConcurrentRequests: 15000,
			ErrorPercentThreshold: 25, RequestVolumeThreshold: 500, SleepWindowInSecond: 10,
		},
		CommandName: "Broker-Command",
		CircuitProneErrors: []string{
			sarama.ErrClosedClient.Error(),
			sarama.ErrOutOfBrokers.Error(),
			sarama.ErrNotConnected.Error(),
			sarama.ErrShuttingDown.Error(),
			sarama.ErrControllerNotAvailable.Error(),
			sarama.ErrNoTopicsToUpdateMetadata.Error(),
			ErrPublishSendMessageTimeout.Error(),
			"You must provide at least one broker address",
		},
	}
}

func (c *Config) timeoutInSecond() time.Duration {
	if c == nil || c.TimeoutInSecond == 0 {
		c.TimeoutInSecond = 3
	}
	return time.Duration(c.TimeoutInSecond) * time.Second
}

func (c *Config) maxMessageBytes() int {
	if c == nil || c.MaxMessageBytes == 0 {
		return 1000
	}
	return c.MaxMessageBytes
}

func (c *Config) reconnectIntervalInSecond() time.Duration {
	if c == nil || c.ReconnectIntervalInSecond == 0 {
		return c.timeoutInSecond() * 4
	}
	return time.Duration(c.ReconnectIntervalInSecond) * time.Second
}

func (c *Config) maxReconnectRetry() int {
	if c == nil || c.MaxReconnectRetry == 0 {
		return 20
	}
	return c.MaxReconnectRetry
}

func (c *Config) cleanupTimeInSecond() time.Duration {
	if c == nil || c.CleanupTimeInSecond == 0 {
		return c.timeoutInSecond() * 5
	}
	return time.Duration(c.CleanupTimeInSecond) * time.Second
}

func (c *Config) circuitBreaker() *circuit.Config {
	if c == nil || c.CircuitBreaker == nil {
		return &circuit.Config{
			Enabled: true, TimeoutInSecond: 1, MaxConcurrentRequests: 15000,
			ErrorPercentThreshold: 25, RequestVolumeThreshold: 500, SleepWindowInSecond: 10,
		}
	}
	return c.CircuitBreaker
}

func (c *Config) commandName() string {
	if c == nil || c.CommandName == "" {
		return "Broker-Command"
	}
	return c.CommandName
}

// CompressionType is a struct to enable or disable log level
type CompressionType struct {
	codec sarama.CompressionCodec
	value string
}

var (
	// None is used for disabling all the Compression
	None = CompressionType{sarama.CompressionNone, "\"NONE\""}
	// GZIP is used for GZIP Compression
	GZIP = CompressionType{sarama.CompressionGZIP, "\"GZIP\""}
	// Snappy is used for Snappy Compression
	Snappy = CompressionType{sarama.CompressionSnappy, "\"SNAPPY\""}
	// LZ4 is used for LZ4 Compression
	LZ4 = CompressionType{sarama.CompressionLZ4, "\"LZ4\""}
)

// UnmarshalJSON is a function to unmarshal CompressionType
// The default value is NONE
func (c *CompressionType) UnmarshalJSON(data []byte) error {
	compression := strings.ToUpper(strings.TrimSpace(string(data)))
	comp := None
	switch compression {
	case None.value:
		comp = None
	case GZIP.value:
		comp = GZIP
	case Snappy.value:
		comp = Snappy
	case LZ4.value:
		comp = LZ4
	}

	c.codec = comp.codec
	c.value = comp.value
	return nil
}

// MarshalJSON is a function to marshal CompressionType
func (c CompressionType) MarshalJSON() ([]byte, error) {
	return []byte(c.value), nil
}
