package producer

import "github.com/confluentinc/confluent-kafka-go/kafka"

// Message : The collection of elements passed to the Producer in order to send a message
type Message struct {
	// Topic: The Kafka topic for this message
	Topic string
	// Key: The partitioning key for this message
	Key []byte
	// Value: The actual message to store in Kafka
	Value []byte

	headers map[string]string
}

// AddHeader : Add header to message
func (m *Message) AddHeader(key string, value string) {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	m.headers[key] = value
}

func (m *Message) toKafkaMessage() *kafka.Message {
	message := &kafka.Message{
		Key:   m.Key,
		Value: m.Value,
		TopicPartition: kafka.TopicPartition{
			Topic:     &m.Topic,
			Partition: kafka.PartitionAny,
		},
	}
	if len(m.headers) > 0 {
		message.Headers = []kafka.Header{}
		for k, v := range m.headers {
			message.Headers = append(message.Headers, kafka.Header{Key: k, Value: []byte(v)})
		}
	}
	return message
}
