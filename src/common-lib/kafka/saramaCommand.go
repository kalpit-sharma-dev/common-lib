// Package kafka implements kafka client configuration details
//
// Deprecated: kafka is old implementation of kafka connectivity and should not be used
// except for compatibility with legacy systems.
//
// Use https://gitlab.kksharmadevdev.com/platform/Platform-Infrastructure-lib/tree/master/messaging for all kafka connectivity
// This package is frozen and no new functionality will be added.
package kafka

import (
	"fmt"
	"log"
	"time"

	"github.com/Shopify/sarama"
	saramacluster "github.com/bsm/sarama-cluster"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/kafka/encode"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

// ProducerCommandImpl implements a ProducerCommand
type saramaProducerCommandImpl struct {
	syncProducer sarama.SyncProducer
}

// NewProducer create a new SaramaProducer
func (pc *saramaProducerCommandImpl) NewProducer(brokerAddress []string) error {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokerAddress, config)
	if err != nil {
		return err
	}
	pc.syncProducer = producer
	return err
}

// Close closes connnection to kafka
func (pc *saramaProducerCommandImpl) Close() error {
	err := pc.syncProducer.Close()
	if err != nil {
		return err
	}
	pc.syncProducer = nil
	return err
}

// PushMessage pushes new Message to kafka
func (pc *saramaProducerCommandImpl) PushMessage(topicName string, message string) (int32, int64, error) {
	// //TODO: Pushed messages may be logged with DEBUG level once logging pattern in implemented in common-lib
	// msg := sarama.ProducerMessage{Topic: topicName, Value: sarama.StringEncoder(message)}
	// p, o, e := pc.syncProducer.SendMessage(&msg)
	// if e != nil {
	// 	fmt.Println("Push Error")
	// 	fmt.Println(e)
	// }
	//pc.syncProducer = nil
	return pc.PushMessageEncoder(topicName, encode.GetStringEncoder(message))
}

// PushMessage pushes new Message to kafka
func (pc *saramaProducerCommandImpl) PushMessageEncoder(topicName string, message encode.Encoder) (int32, int64, error) {
	//TODO: Pushed messages may be logged with DEBUG level once logging pattern in implemented in common-lib
	log := logger.Get()
	msg := sarama.ProducerMessage{Topic: topicName, Value: message}
	partition, offset, err := pc.syncProducer.SendMessage(&msg)
	if err != nil {
		log.Error("", "PushMessageEncoder:SendMessage", "Error Producing Message: Error %v", err)
		return 0, 0, err

	}

	//pc.syncProducer = nil
	return partition, offset, nil
}

// IsConnected checks if the Producer Connected
func (pc *saramaProducerCommandImpl) IsConnected() bool {
	if pc.syncProducer == nil {
		return false
	}
	return true
}

// saramaConsumerCommandImpl implements ConsumerCommand
type saramaConsumerCommandImpl struct {
	config   *sarama.Config
	consumer *saramacluster.Consumer
}

func (cc *saramaConsumerCommandImpl) NewConsumer(brokerAddress []string, GroupID string, Topics []string) error {
	config := saramacluster.NewConfig()
	config.Consumer.Offsets.Retention = 1
	consumer, err := saramacluster.NewConsumer(brokerAddress, GroupID, Topics, config)
	if err != nil {
		log.Println(err)
		return err
	}
	cc.consumer = consumer
	return err
}

func (cc *saramaConsumerCommandImpl) NewConsumerSafe(brokerAddress []string,
	GroupID string, Topics []string) error {
	config := saramacluster.NewConfig()
	config.Consumer.Offsets.Retention = 1
	config.Consumer.Offsets.CommitInterval = 720 * time.Hour
	consumer, err := saramacluster.NewConsumer(brokerAddress, GroupID, Topics, config)
	if err != nil {
		log.Println(err)
		return err
	}
	cc.consumer = consumer
	return err
}

func (cc *saramaConsumerCommandImpl) NewCustomConsumer(
	inOut *ConsumerKafkaInOutParams, brokerAddress []string,
	GroupID string, Topics []string) error {

	config := saramacluster.NewConfig()
	config.Consumer.Return.Errors = inOut.ReturnErrors
	config.Group.Return.Notifications = inOut.ReturnNotifications
	config.Consumer.Offsets.Initial = inOut.OffsetsInitial
	config.Consumer.Offsets.Retention = inOut.Retention
	config.Version = sarama.V1_1_0_0
	consumer, err := saramacluster.NewConsumer(brokerAddress, GroupID, Topics, config)
	if err == nil {
		cc.consumer = consumer

		// Process Kafka cluster errors
		go func(consumer *saramacluster.Consumer) {
			for err := range consumer.Errors() {
				inOut.Errors <- err
			}

		}(consumer)

		// Process Kafka cluster notifications
		go func(consumer *saramacluster.Consumer) {
			for ntf := range consumer.Notifications() {
				inOut.Notifications <- fmt.Sprintf("%+v", ntf)
			}

		}(consumer)

	}
	return err
}

func (cc *saramaConsumerCommandImpl) IsConnected() bool {
	if cc.consumer == nil {
		return false
	}
	return true
}

func (cc *saramaConsumerCommandImpl) Close() error {
	err := cc.consumer.Close()
	if err != nil {
		return err
	}
	cc.consumer = nil
	return err
}

func (cc *saramaConsumerCommandImpl) PullMessage(consumerHandler ConsumerHandler) {
	chmsg := cc.consumer.Messages()
	for {
		message := <-chmsg
		if message == nil {
			log.Printf("PullMessage - Received msg %v. Channel closed.", message)
			break
		}
		cc.consumer.MarkPartitionOffset(message.Topic, message.Partition, message.Offset, "")
		consumerMessage := ConsumerMessage{Message: string(message.Value), Offset: message.Offset, Partition: message.Partition, Topic: message.Topic, ReceivedDateTimeUTC: time.Now().UTC()}
		//TODO: Pulled messages may be logged with DEBUG level once logging pattern in implemented in common-lib
		go consumerHandler(consumerMessage)
	}
}

func (cc *saramaConsumerCommandImpl) PullMessageProcessSequentially(consumerHandler ConsumerHandler) {
	chmsg := cc.consumer.Messages()
	for {
		message := <-chmsg
		if message == nil {
			log.Printf("PullMessageProcessSequentially - Received msg %v. Channel closed.", message)
			break
		}

		consumerMessage := ConsumerMessage{Message: string(message.Value), Offset: message.Offset, Partition: message.Partition, Topic: message.Topic, ReceivedDateTimeUTC: time.Now().UTC()}
		consumerHandler(consumerMessage)
		cc.consumer.MarkPartitionOffset(message.Topic, message.Partition, message.Offset, "")
		cc.consumer.CommitOffsets()
	}
}

func (cc *saramaConsumerCommandImpl) LimitedPullMessageNoOffset(consumerHandler ConsumerHandler, limiter Limiter) {

	for {
		if limiter.IsConsumingAllowed() {
			message := <-cc.consumer.Messages()
			consumerMessage := ConsumerMessage{
				Message:             string(message.Value),
				Offset:              message.Offset,
				Partition:           message.Partition,
				Topic:               message.Topic,
				ReceivedDateTimeUTC: time.Now().UTC(),
			}
			consumerHandler(consumerMessage)
		} else {
			limiter.Wait()
		}
	}
}

func (cc *saramaConsumerCommandImpl) MarkOffset(topic string, partition int32, offset int64) {
	cc.consumer.MarkPartitionOffset(topic, partition, offset, "")
}
