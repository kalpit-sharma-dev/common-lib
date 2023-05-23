package publisher

import (
	"context"
	"runtime/debug"
	"strings"

	"github.com/Shopify/sarama"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

// SyncProducer - publish messages using Kafka Sync producer
type syncProducer struct {
	cfg          *Config
	producerType ProducerType
	producer     sarama.SyncProducer
	existing     sarama.SyncProducer
}

// syncProducer is a function to return instance of sarama sync producer
func (s *syncProducer) syncProducer(transaction string) (sarama.SyncProducer, error) {
	producerMutex.RLock()
	if s.producer == nil {
		producerMutex.RUnlock()
		producerMutex.Lock()
		defer producerMutex.Unlock()
		if s.producer == nil {
			prod, err := sarama.NewSyncProducer(s.cfg.Address, GetConfig(s.producerType, s.cfg))
			if err != nil {
				Logger().Error(transaction, "sync.producer.creation.failed", "Error in creating %s Kafka sync producer %v", s.producerType, err)
				return nil, err
			}
			s.producer = prod
		}
	} else {
		defer producerMutex.RUnlock()
	}
	return s.producer, nil
}

// Publish - publish kafka messages with context
func (s *syncProducer) Publish(ctx context.Context, transaction string, messages ...*Message) error {
	if len(messages) == 0 {
		return ErrPublishMessageNotAvailable
	}

	var err error

	circuitErr := circuit.Do(s.cfg.commandName(), s.cfg.circuitBreaker().Enabled, func() error {
		err = s.publish(ctx, transaction, messages...)
		if s.validCircuitError(err) {
			return err
		}
		return nil
	}, nil)

	if circuitErr != nil {
		return circuitErr
	}

	return err
}

// We should open CB only in case of legitimate Kafka issues.
// LEADER_NOT_AVAILABLE 5, REQUEST_TIMED_OUT 7, BROKER_NOT_AVAILABLE 8
// The actual error error would be stored in the closure.
// The call back handler will check for error code returned by Kafka and if its not 5, 7 and 8
// then CB should be opened otherwise it should be closed.
// At the same time, caller would be able to receive the actual error code.
func (s *syncProducer) validCircuitError(err error) bool {
	if len(s.cfg.CircuitProneErrors) == 0 {
		return true
	}

	if err == nil {
		return false
	}

	if s.circuitProneErrors(err) {
		return true
	}

	// The actual error error would be stored in the closure.
	// The call back handler will check for error code returned by Kafka and if its not 5, 7 and 8
	// then CB should be opened otherwise it should be closed.
	// At the same time, caller would be able to receive the actual error code.
	pes, ok := err.(sarama.ProducerErrors)
	if ok {
		for _, pe := range pes {
			if pe != nil && s.circuitProneErrors(pe) {
				return true
			}
		}
	}
	return false
}

func (s *syncProducer) circuitProneErrors(err error) bool {
	for _, v := range s.cfg.CircuitProneErrors {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}
	return false
}

func (s *syncProducer) publish(ctx context.Context, transaction string, messages ...*Message) error {
	errChan := make(chan error, 1)
	go func(errChan chan error) {
		defer func() {
			if r := recover(); r != nil {
				Logger().Error(transaction, "publish:SendMessage:Recovered", "Recovered in Publish %v and Trace : %s", r, debug.Stack())
				errChan <- ErrPublishSendMessageRecovered
			}
		}()

		p, err := s.syncProducer(transaction)
		if err != nil {
			errChan <- err
			return
		}

		msgs := make([]*sarama.ProducerMessage, len(messages))

		for index, message := range messages {
			msgs[index] = &sarama.ProducerMessage{
				Topic:   message.Topic,
				Value:   message.Value,
				Key:     message.Key,
				Headers: message.toRecordHeader(transaction),
			}
		}

		err = p.SendMessages(msgs)
		errChan <- err
	}(errChan)

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ErrPublishSendMessageTimeout
	}
}

// reconnect - reconnect to sync producer
func (s *syncProducer) reconnect(transaction string) error {
	s.existing = s.producer
	prod, err := sarama.NewSyncProducer(s.cfg.Address, GetConfig(s.producerType, s.cfg))
	if err != nil {
		return err
	}
	s.producer = prod
	return nil
}

// cleanExisting - clean existing to sync producer
func (s *syncProducer) cleanExisting(transaction string) error {
	if s.existing != nil {
		return s.existing.Close()
	}
	return nil
}

// clean - clean sync producer
func (s *syncProducer) clean(transaction string) error {
	if s.producer != nil {
		return s.producer.Close()
	}
	return nil
}

// Utility method to extract the actual
// sarama error msg strings from sarama.ProducerErrors,
// collate them and return that the consumer can log the exact reason
// for failure.
func GetSaramaErrorMsg(err error) string {
	if pErrs, ok := err.(sarama.ProducerErrors); ok {
		var saramaErrorMsgs strings.Builder
		for _, pErr := range pErrs {
			saramaErrorMsgs.WriteString(pErr.Error() + "\n")
		}
		return saramaErrorMsgs.String()
	}
	return ""

}
