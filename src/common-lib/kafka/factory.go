// Package kafka implements kafka client configuration details
//
// Deprecated: kafka is old implementation of kafka connectivity and should not be used
// except for compatibility with legacy systems.
//
// Use https://gitlab.kksharmadevdev.com/platform/Platform-Infrastructure-lib/tree/master/messaging for all kafka connectivity
// This package is frozen and no new functionality will be added.
package kafka

import (
	"github.com/Shopify/sarama"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/rest"
)

// ProducerFactoryImpl implments ProducerFactory
type ProducerFactoryImpl struct {
}

// GetProducerService returns a ProducerService
func (ProducerFactoryImpl) GetProducerService(config ProducerConfig) (ProducerService, error) {
	cmdFactory := new(ProducerCommandFactoryImpl)
	return newSaramaProducer(&config, cmdFactory)
}

// GetConfluentProducerService gets confluent Producer
func (ProducerFactoryImpl) GetConfluentProducerService(config ProducerConfig) (ProducerService, error) {
	cmdFactory := new(ProducerCommandFactoryImpl)
	return newSaramaProducer(&config, cmdFactory)
}

// ConsumerFactoryImpl returns a ConsumerConfig
type ConsumerFactoryImpl struct {
}

// GetConsumerService return a ConsumerService
func (ConsumerFactoryImpl) GetConsumerService(config ConsumerConfig) (ConsumerService, error) {
	cmdFactory := new(ConsumerCommandFactoryImpl)
	return newSaramaConsumer(&config, cmdFactory)
}

// ProducerCommandFactoryImpl implements Factory mathod that gets ProducerCommandService
type ProducerCommandFactoryImpl struct {
}

// GetProducerCommandService returns implementation of ProducerCommand
func (ProducerCommandFactoryImpl) GetProducerCommandService() ProducerCommand {
	return new(saramaProducerCommandImpl)
}

// GetConfluentProducerCommandService return confluent implementation of producer command
func (ProducerCommandFactoryImpl) GetConfluentProducerCommandService() ProducerCommand {
	return new(saramaProducerCommandImpl)
}

// ConsumerCommandFactoryImpl implements Factory mathod that gets ProducerCommandService
type ConsumerCommandFactoryImpl struct {
}

// GetConsumerCommandService returns implementation of ConsumerCommand
func (ConsumerCommandFactoryImpl) GetConsumerCommandService() ConsumerCommand {
	return new(saramaConsumerCommandImpl)
}

// GetConsumerCommandServiceSafe returns implementation of CosumerCommand
func (ConsumerCommandFactoryImpl) GetConsumerCommandServiceSafe() ConsumerCommandSequential {
	return new(saramaConsumerCommandImpl)
}

// Health returns a Health state for Kafka
func (ProducerFactoryImpl) Health(kafkaBrokers []string) rest.Statuser {
	cmdFactory := new(ProducerCommandFactoryImpl)
	return status{
		kafkaBrokers: kafkaBrokers,
		factory:      cmdFactory,
	}
}

type status struct {
	kafkaBrokers []string
	factory      ProducerCommandFactory
}

func (k status) Status(conn rest.OutboundConnectionStatus) *rest.OutboundConnectionStatus {
	conn.ConnectionType = "Kafka"
	conn.ConnectionURLs = k.kafkaBrokers
	config := sarama.NewConfig()
	config.Version = sarama.V0_10_0_0
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(k.kafkaBrokers, config)
	if err != nil {
		conn.ConnectionStatus = rest.ConnectionStatusUnavailable
		return &conn
	}

	producer.Close()
	conn.ConnectionStatus = rest.ConnectionStatusActive
	return &conn
}
