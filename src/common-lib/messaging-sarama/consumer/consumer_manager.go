package consumer

import (
	"fmt"
	"runtime"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

type consumerStrategy interface {
	pull(sc *saramaConsumer)
}

func newMessage(message *sarama.ConsumerMessage) Message {
	consumerMessage := Message{
		Message:           message.Value,
		Offset:            message.Offset,
		Partition:         message.Partition,
		Topic:             message.Topic,
		PulledDateTimeUTC: time.Now().UTC(),
		headers:           make(map[string]string),
	}

	for _, h := range message.Headers {
		consumerMessage.headers[string(h.Key)] = string(h.Value)
	}

	return consumerMessage
}

func getConsumerStrategy(cfg Config) consumerStrategy {
	Logger().Trace(cfg.TransactionID, "Consumer Strategy for Consumer Mode: %v", cfg.ConsumerMode)
	switch cfg.ConsumerMode {
	case PullOrdered:
		return &pullOrdered{cfg: cfg}
	case PullUnOrdered:
		return &pullUnOrdered{cfg: cfg}
	case PullOrderedWithOffsetReplay:
		return getNewpullOrderedWithOffsetReplay(cfg)
	}

	panic(fmt.Errorf("consumer Strategy not available for Consumer Mode: %v", cfg.ConsumerMode))
}

//PullUnOrdered Stratgy Handler
type pullUnOrdered struct {
	cfg Config
}

func (o *pullUnOrdered) pull(sc *saramaConsumer) {
	for message := range sc.consumer.Messages() {
		sc.canConsume.Store(true)
		consumerMessage := newMessage(message)
		sc.commitStrategy.onPull(consumerMessage.GetTransactionID(), message.Topic, message.Partition, message.Offset)
		sc.pool.addJob(consumerMessage, sc.commitStrategy)
	}
}

//PullOrdered - Stratgy Handler
type pullOrdered struct {
	cfg Config
}

func (o *pullOrdered) processMessage(consumerMessage Message, sc *saramaConsumer) {
	var retryCount int64

	sc.commitStrategy.beforeHandler(consumerMessage.GetTransactionID(), consumerMessage.Topic,
		consumerMessage.Partition, consumerMessage.Offset)

	var err error
	for retryCount = 0; retryCount < sc.cfg.RetryCount; retryCount++ {
		err = invokeMessageHandler(consumerMessage, sc.cfg)
		if err == nil {
			break
		}
		time.Sleep(sc.cfg.RetryDelay)
	}
	if err != nil {
		invokeErrorHandler(err, &consumerMessage, sc.cfg)
	}

	sc.commitStrategy.afterHandler(consumerMessage.GetTransactionID(), consumerMessage.Topic,
		consumerMessage.Partition, consumerMessage.Offset)
}

//nolint:gosimple
func (o *pullOrdered) pull(sc *saramaConsumer) {
	log := Logger() // nolint - false positive
	log.Trace(o.cfg.TransactionID, "Starting Consumer")
	routine := make(chan cluster.PartitionConsumer, o.cfg.Partitions)
	poolSize := sc.cfg.SubscriberPerCore * runtime.NumCPU()

	for index := 0; index < poolSize; index++ {
		log.Trace(o.cfg.TransactionID, "Starting Consumer routine : %v", index)
		go o.consume(sc, routine)
	}

	// consume partitions
	for {
		select {
		case part, ok := <-sc.consumer.Partitions():
			if !ok {
				return
			}

			select {
			case routine <- part:
				log.Trace(o.cfg.TransactionID, "Sending Topic :: %s, Partition :: %v for Consumption", part.Topic(), part.Partition())
			default:
				log.Info(o.cfg.TransactionID, "No routine available for Partition :: %v and Topic :: %s", part.Partition(), part.Topic())
			}
		}
	}
}

// start a separate goroutine to consume messages
func (o *pullOrdered) consume(sc *saramaConsumer, routine chan cluster.PartitionConsumer) {
	for {
		pc := <-routine
		log := Logger() // nolint - false positive
		log.Debug(o.cfg.TransactionID, "Starting Consumer Routine for Topic :: %s and Partition :: %v", pc.Topic(), pc.Partition())
		select {
		case msg, ok := <-pc.Messages():
			if ok {
				consumerMessage := newMessage(msg)
				sc.commitStrategy.onPull(consumerMessage.GetTransactionID(), msg.Topic, msg.Partition, msg.Offset)
				o.processMessage(consumerMessage, sc)
				routine <- pc
			} else {
				log.Debug(o.cfg.TransactionID, "Removing listener for Topic :: %s and Partition :: %v", pc.Topic(), pc.Partition())
			}
		default:
			log.Debug(o.cfg.TransactionID, "Moving to next Topic :: %s and Partition :: %v", pc.Topic(), pc.Partition())
			routine <- pc
			// Added this sleep to avoid CPU spike, if there is nothing to consume in the partitions
			time.Sleep(time.Second)
		}
	}
}
