// Package kafka implements kafka client configuration details
//
// Deprecated: kafka is old implementation of kafka connectivity and should not be used
// except for compatibility with legacy systems.
//
// Use https://gitlab.kksharmadevdev.com/platform/Platform-Infrastructure-lib/tree/master/messaging for all kafka connectivity
// This package is frozen and no new functionality will be added.
package kafka

import (
	"errors"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/kafka/encode"
)

type saramaProducerImpl struct {
	config  *ProducerConfig
	command ProducerCommand
}

// newSaramaProducer returns new instance of SaramaProducer with the provided configuration
func newSaramaProducer(config *ProducerConfig, cmdFactory ProducerCommandFactory) (*saramaProducerImpl, error) {
	producer := new(saramaProducerImpl)
	err := validateConfig(&config.ClientConfig)

	if err != nil {
		return nil, err
	}
	producer.config = config
	producer.command = cmdFactory.GetProducerCommandService()
	return producer, nil
}

// newSaramaProducer returns new instance of SaramaProducer with the provided configuration
func newConfluentProducer(config *ProducerConfig, cmdFactory ProducerCommandFactory) (*saramaProducerImpl, error) {
	producer := new(saramaProducerImpl)
	err := validateConfig(&config.ClientConfig)

	if err != nil {
		return nil, err
	}
	producer.config = config
	producer.command = cmdFactory.GetConfluentProducerCommandService()
	return producer, nil
}

// Connect connects the producer to the kafka broker
func (sp *saramaProducerImpl) connect() error {

	result := sp.isConnected()
	if result == true {
		return nil
	}
	err := sp.command.NewProducer(sp.config.BrokerAddress)

	return err
}

// Push pushes a new message to the specified kafka topic
func (sp *saramaProducerImpl) Push(topicName string, message string) error {
	var err error

	err = sp.connect()
	if err != nil {
		return err
	}

	var result bool
	result = sp.isConnected()
	if result == false {
		return errors.New(ErrorClientNotConnected)
	}

	_, _, err = sp.command.PushMessage(topicName, message)
	if err != nil {
		return err
	}
	return nil
}

func (sp *saramaProducerImpl) PushEncoder(topicName string, message encode.Encoder) (err error) {
	err = sp.connect()
	if err != nil {
		return err
	}

	var result bool
	result = sp.isConnected()
	if result == false {
		return errors.New(ErrorClientNotConnected)
	}

	_, _, err = sp.command.PushMessageEncoder(topicName, message)
	if err != nil {
		return err
	}
	return nil
}

// isConnected checks if the Producer is connected to kafka server
func (sp *saramaProducerImpl) isConnected() bool {
	return sp.command.IsConnected()
}

// CloseConnection closes Producers connection to kafka server
func (sp *saramaProducerImpl) CloseConnection() error {
	var result bool

	result = sp.command.IsConnected()
	if result == false {
		return nil
	}

	err := sp.command.Close()
	if err != nil {
		return err
	}
	return nil
}
