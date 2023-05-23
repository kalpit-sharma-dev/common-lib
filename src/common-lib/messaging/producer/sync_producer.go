package producer

import (
	"context"
	"github.com/pkg/errors"
	"runtime/debug"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type syncProducer struct {
	producer   *kafka.Producer
	fatalError error
	health     *Health
}

func (sp *syncProducer) Produce(ctx context.Context, transaction string, messages ...*Message) error {
	_, err := sp.ProduceWithReport(ctx, transaction, messages...)
	return err
}

func (sp *syncProducer) ProduceWithReport(ctx context.Context, transaction string, messages ...*Message) ([]*DeliveryReport, error) {
	if err := sp.producer.GetFatalError(); err != nil {
		return nil, errors.WithMessage(ErrPublishSendMessageFatal, err.Error())
	}
	if len(messages) == 0 {
		return nil, ErrPublishMessageNotAvailable
	}
	resultChan := make(chan produceResult, 1)
	go func(resultChan chan produceResult) {
		defer func() {
			if r := recover(); r != nil {
				Logger().Error(transaction, "publish:SendMessage:Recovered", "Recovered in Publish %v and Trace : %s", r, debug.Stack())
				resultChan <- produceResult{nil, ErrPublishSendMessageRecovered}
			}
		}()

		deliveryChan := make(chan kafka.Event, len(messages))

		deliveryReports := make([]*DeliveryReport, 0, len(messages))
		var failedReports []DeliveryReport

		var errCount = 0
		for _, message := range messages {
			msg := message.toKafkaMessage()
			msg.Opaque = message
			err := sp.producer.Produce(msg, deliveryChan)
			if err != nil {
				errCount++
				dr := DeliveryReport{
					Message: message,
					Error:   err,
				}
				deliveryReports = append(deliveryReports, &dr)
				failedReports = append(failedReports, dr)
			}
		}

		for i := 0; i < len(messages)-errCount; i++ {
			e := <-deliveryChan
			m, ok := e.(*kafka.Message)
			if !ok {
				continue
			}
			dr := DeliveryReport{}
			dr.Message = m.Opaque.(*Message)
			if m.TopicPartition.Error != nil {
				dr.Error = m.TopicPartition.Error
				failedReports = append(failedReports, dr)
			} else {
				dr.TopicPartition = TopicPartition{
					Topic:     *m.TopicPartition.Topic,
					Partition: m.TopicPartition.Partition,
					Offset:    int64(m.TopicPartition.Offset),
				}
			}
			deliveryReports = append(deliveryReports, &dr)
		}

		close(deliveryChan)

		if len(failedReports) > 0 {
			resultChan <- produceResult{deliveryReports, &DeliveryError{FailedReports: failedReports}}
			return
		}
		resultChan <- produceResult{deliveryReports, nil}
	}(resultChan)

	select {
	case result := <-resultChan:
		return result.deliveryReport, result.err
	case <-ctx.Done():
		return nil, ErrPublishSendMessageTimeout
	}
}

func (sp *syncProducer) Close() {
	sp.producer.Close()
}

func (sp *syncProducer) Health() (*Health, error) {
	return sp.health, nil
}

func (sp *syncProducer) GetFatalError() error {
	return sp.producer.GetFatalError()
}

func (sp *syncProducer) connStateChange(state bool) {
	sp.health.ConnectionState = state
}

// NewSyncProducer returns a synchronous implementation of StatusableProducer.
var NewSyncProducer = func(config *Config) (StatusableProducer, error) {
	sp := &syncProducer{
		health: newHealth(),
	}
	sp.health.Address = config.Address
	sp.health.ConnectionState = true
	p, err := newKafkaProducer(config, sp.connStateChange)
	if err != nil {
		return nil, err
	}
	sp.producer = p
	return sp, nil
}
