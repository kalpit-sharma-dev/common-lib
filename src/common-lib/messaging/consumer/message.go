package consumer

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/constants"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
)

// Message : Represents consumed Kafka Message
type Message struct {
	Key               []byte
	Message           []byte
	Offset            int64
	Partition         int32
	Topic             string
	PulledDateTimeUTC time.Time
	headers           map[string]string
	transactionID     string
}

// GetHeaders - Retrieve All headers
func (m *Message) GetHeaders() map[string]string {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	newHeaders := make(map[string]string)
	for k, v := range m.headers {
		newHeaders[k] = v
	}
	return newHeaders
}

// GetHeader - Retrieve a value from the message headers
func (m *Message) GetHeader(key string) string {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	return m.headers[key]
}

// GetTransactionID - Retrieve Transaction ID value from the message headers
func (m *Message) GetTransactionID() string {
	if m.transactionID == "" {
		m.transactionID = utils.GetTransactionID()
		if m.headers != nil {
			transaction, ok := m.headers[constants.TransactionID]
			if ok {
				m.transactionID = transaction
			}
		}
	}
	return m.transactionID
}

func newMessage(message *kafka.Message) *Message {
	consumerMessage := &Message{
		Key:               message.Key,
		Message:           message.Value,
		Offset:            int64(message.TopicPartition.Offset),
		Partition:         message.TopicPartition.Partition,
		Topic:             *message.TopicPartition.Topic,
		PulledDateTimeUTC: time.Now().UTC(),
		headers:           make(map[string]string),
	}

	for _, h := range message.Headers {
		consumerMessage.headers[string(h.Key)] = string(h.Value)
	}

	consumerMessage.GetTransactionID()

	return consumerMessage
}
