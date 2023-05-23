package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"hash"

	"github.com/OneOfOne/xxhash"
	"github.com/Shopify/sarama"
)

// Producer ...
type Producer interface {
	// Publish - publish kafka messages with context
	Publish(ctx context.Context, transaction string, messages ...*Message) error
}

type producer interface {
	// Publish - publish kafka messages with context
	Publish(ctx context.Context, transaction string, messages ...*Message) error
	reconnect(transaction string) error
	cleanExisting(transaction string) error
	clean(transaction string) error
}

// Encoder is a simple interface for any type that can be encoded as an array of bytes
// in order to be sent as the key or value of a Kafka message. Length() is provided as an
// optimization, and must return the same as len() on the result of Encode().
type Encoder interface {
	sarama.Encoder
}

//ProducerType denotes kafka producer to be used
type ProducerType string

const (
	//RegularKafkaProducer denotes normal kafka producer
	RegularKafkaProducer ProducerType = "regular"

	//BigKafkaProducer denotes kafka producer capable of publishing big messages
	BigKafkaProducer ProducerType = "big"
)

// Error Codes
var (
	// ErrPublishMessageNotAvailable : Error message not available
	ErrPublishMessageNotAvailable = errors.New("Publish:Message:Not.Available")

	// ErrPublishSendMessageRecovered : Error publish message thrown panic
	ErrPublishSendMessageRecovered = errors.New("Publish:SendMessage:Recovered")

	// ErrPublishSendMessageTimeout : Publish message timed-out
	ErrPublishSendMessageTimeout = errors.New("Publish:SendMessage:Timeout")
)

// Message is the collection of elements passed to the Producer in order to send a message.
type Message struct {
	// Topic - The Kafka topic for this message.
	Topic string
	// Key - The partitioning key for this message. Pre-existing Encoders include
	// StringEncoder and ByteEncoder.
	Key Encoder
	// Value - The actual message to store in Kafka. Pre-existing Encoders include
	// StringEncoder and ByteEncoder.
	Value Encoder

	// Headers - The headers are key-value pairs that are transparently passed
	// by Kafka between producers and consumers.
	headers map[string]string
}

// AddHeader - The headers are key-value pairs that are transparently passed
func (m *Message) AddHeader(key string, value string) {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	m.headers[key] = value
}

// AddHeader - The headers are key-value pairs that are transparently passed
func (m *Message) toRecordHeader(transaction string) []sarama.RecordHeader {
	headers := make([]sarama.RecordHeader, 0)

	if m.headers != nil {
		for key, value := range m.headers {
			headers = append(headers, sarama.RecordHeader{Key: []byte(key), Value: []byte(value)})
		}
	}

	return headers
}

//GetConfig returns a Sarama Config for Kafka
func GetConfig(producerType ProducerType, cfg *Config) *sarama.Config {
	// For the data collector, we are looking for strong consistency semantics.
	// Because we don't change the flush settings, sarama will try to produce messages
	// as fast as possible to keep latency low.
	config := sarama.NewConfig()
	config.Metadata.Full = false
	config.Producer.Return.Successes = true
	config.Version = sarama.MaxVersion
	config.Producer.Timeout = cfg.timeoutInSecond()

	config.Producer.Partitioner = sarama.NewCustomHashPartitioner(func() hash.Hash32 {
		return xxhash.New32()
	})

	if producerType == BigKafkaProducer {
		config.Producer.MaxMessageBytes = cfg.maxMessageBytes()
		config.Producer.Compression = cfg.CompressionType.codec
	}

	return config
}

//EncodeString is a function to encode String for Sarama
func EncodeString(value string) Encoder {
	return sarama.StringEncoder(value)
}

//EncodeBytes is a function to encode bytes for Sarama
func EncodeBytes(value []byte) Encoder {
	return sarama.ByteEncoder(value)
}

//EncodeObject is a function to encode object for Sarama
var EncodeObject = func(value interface{}) (Encoder, error) {
	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(value)
	if err != nil {
		return nil, err
	}
	return sarama.ByteEncoder(buffer.Bytes()), nil
}
