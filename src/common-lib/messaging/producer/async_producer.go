package producer

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type asyncProducer struct {
	producer           *kafka.Producer
	produceChan        chan *Message
	deliveryReportChan chan *DeliveryReport
	health             *Health
}

func (ap *asyncProducer) ProduceChannel() chan<- *Message {
	return ap.produceChan
}

func (ap *asyncProducer) DeliveryReportChannel() <-chan *DeliveryReport {
	return ap.deliveryReportChan
}

func (ap *asyncProducer) Flush(timeoutMs int) int {
	return ap.producer.Flush(timeoutMs)
}

func (ap *asyncProducer) Close() {
	close(ap.produceChan)
}

func (ap *asyncProducer) Health() (*Health, error) {
	return ap.health, nil
}

func (ap *asyncProducer) connStateChange(state bool) {
	ap.health.ConnectionState = state
}

func (ap *asyncProducer) processMessages(pc chan *Message) {
	for message := range pc {
		msg := message.toKafkaMessage()
		msg.Opaque = message
		ap.producer.ProduceChannel() <- msg
	}
	ap.producer.Close()
}

func (ap *asyncProducer) processEvents(ec chan kafka.Event) {
	for ev := range ec {
		m, ok := ev.(*kafka.Message)
		if !ok {
			continue
		}
		dr := &DeliveryReport{}
		dr.Message, _ = m.Opaque.(*Message)
		if m.TopicPartition.Error != nil {
			dr.Error = m.TopicPartition.Error
		} else {
			dr.TopicPartition = TopicPartition{
				Topic:     *m.TopicPartition.Topic,
				Partition: m.TopicPartition.Partition,
				Offset:    int64(m.TopicPartition.Offset),
			}
		}
		ap.deliveryReportChan <- dr
	}
	close(ap.deliveryReportChan)
}

// NewAsyncProducer returns an async inplementation of Producer
func NewAsyncProducer(config *Config) (AsyncProducer, error) {
	ap := &asyncProducer{
		health:             newHealth(),
		produceChan:        make(chan *Message, config.ProduceChannelSize),
		deliveryReportChan: make(chan *DeliveryReport, config.ProduceChannelSize),
	}
	ap.health.Address = config.Address
	ap.health.ConnectionState = true // asume connection will work, this will change to false if it fails to connect
	p, err := newKafkaProducer(config, ap.connStateChange)
	if err != nil {
		return nil, err
	}
	ap.producer = p
	go ap.processMessages(ap.produceChan)
	go ap.processEvents(ap.producer.Events())
	return ap, nil
}
