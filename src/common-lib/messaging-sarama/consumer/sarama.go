// Package consumer provides a Kafka consumer built on sarama.
//
// Deprecated: use messaging/consumer instead of infra-lib/messaging/consumer
package consumer

import (
	"fmt"
	"runtime/debug"

	"go.uber.org/atomic"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

const returnErrors = true

type saramaLogger struct {
	log logger.Log
}

func (s saramaLogger) Print(v ...interface{}) {
	s.log.Info("Sarama", "%+v", v)
}

func (s saramaLogger) Printf(format string, v ...interface{}) {
	s.log.Info("Sarama", format, v...)
}

func (s saramaLogger) Println(v ...interface{}) {
	s.log.Info("Sarama", "%+v", v)
}

// New function to return Consumer Service
func New(cfg Config) (Service, error) {
	if len(cfg.Address) == 0 || cfg.Group == "" || len(cfg.Topics) == 0 || cfg.MessageHandler == nil {
		return nil, fmt.Errorf("some of the values are blank. MessageHandler, Address : %v, Group : %s and Topics : %v",
			cfg.Address, cfg.Group, cfg.Topics)
	}

	if cfg.SubscriberPerCore <= 0 {
		return nil, fmt.Errorf("failed to create worker pool with SubscriberPerCore %d", cfg.SubscriberPerCore)
	}

	config := cluster.NewConfig()

	config.Consumer.Offsets.Retention = cfg.Retention
	config.Consumer.Return.Errors = returnErrors
	config.Group.Return.Notifications = cfg.NotificationHandler != nil
	config.Consumer.Offsets.Initial = cfg.OffsetsInitial
	config.Metadata.Full = cfg.MetadataFull
	config.Consumer.MaxProcessingTime = cfg.Timeout

	config.ClientID = "Continuum"
	if cfg.KafkaLogConfig != nil {
		log, err := logger.Create(logger.Config{Name: cfg.KafkaLogConfig.FileName, LogLevel: cfg.KafkaLogConfig.LogLevel,
			MaxSize:    cfg.KafkaLogConfig.MaxSize,
			MaxBackups: cfg.KafkaLogConfig.MaxBackups})
		if err == nil {
			sarama.Logger = saramaLogger{
				log: log,
			}
		}
	}
	config.Version = sarama.MaxVersion
	if cfg.ConsumerMode == PullOrdered || cfg.ConsumerMode == PullOrderedWithOffsetReplay {
		config.Group.Mode = cluster.ConsumerModePartitions
	}

	consumer, err := cluster.NewConsumer(cfg.Address, cfg.Group, cfg.Topics, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer : %+v", err)
	}

	sc := &saramaConsumer{
		cfg:        cfg,
		consumer:   consumer,
		canConsume: atomic.NewBool(true),
	}

	if cfg.ConsumerMode == PullUnOrdered {
		pool := &workerPool{cfg: cfg}
		pool.initialize()
		sc.pool = pool
	}

	sc.commitStrategy = getCommitStrategy(cfg, sc.MarkOffset)
	sc.consumerStrategy = getConsumerStrategy(cfg)
	return sc, nil
}

type saramaConsumer struct {
	cfg              Config
	consumer         *cluster.Consumer
	pool             *workerPool
	commitStrategy   commitStrategy
	consumerStrategy consumerStrategy
	canConsume       *atomic.Bool
}

// Pull batch of messages from Kafka topic
func (sc *saramaConsumer) Pull() {
	go sc.consumeError()
	go sc.consumeRebalanceNotification()

	sc.consumerStrategy.pull(sc)
}

func (sc *saramaConsumer) MarkOffset(topic string, partition int32, offset int64) {
	sc.consumer.MarkPartitionOffset(topic, partition, offset, "")
}

func (sc *saramaConsumer) Close() error {
	err := sc.consumer.Close()
	if err != nil {
		return fmt.Errorf("failed to close consumer : %+v", err)
	}

	if sc.pool != nil {
		sc.pool.Shutdown()
	}

	return nil
}

func (sc *saramaConsumer) consumeError() {
	for err := range sc.consumer.Errors() {
		invokeErrorHandler(err, nil, sc.cfg)
		sc.canConsume.Store(false)
	}
}

// Handles notification in case of rebalance
func (sc *saramaConsumer) consumeRebalanceNotification() {
	if sc.cfg.NotificationHandler != nil {
		for notification := range sc.consumer.Notifications() {
			sc.handleConsumerStrategy(notification)
			invokeNotificationHandler(fmt.Sprintf("%+v", notification), sc.cfg)
		}
	}
}

// Executes Consumer mode specific handles for Rebalancing
func (sc *saramaConsumer) handleConsumerStrategy(notification *cluster.Notification) {
	switch sc.cfg.ConsumerMode {
	case PullOrderedWithOffsetReplay:
		cs := sc.consumerStrategy.(*pullOrderedWithOffsetReplay)
		//refresh the pullOrderedWithOffsetReplay partitionkeystore with the
		//new set of partitions which the consumer needs to handle when rebalance is done successfully.
		cs.refreshPartitionKeyStore(notification)
	}

}

func invokeErrorHandler(err error, message *Message, cfg Config) {
	transaction := utils.GetTransactionID()
	if message != nil {
		transaction = message.GetTransactionID()
	}

	defer func() {
		if r := recover(); r != nil {
			Logger().Error(transaction, "invokeErrorHandlerRecovered", "Panic: While processing %v, trace : %s", r, string(debug.Stack()))
		}
	}()
	if cfg.ErrorHandler != nil {
		cfg.ErrorHandler(err, message)
	} else {
		Logger().Debug(transaction, "Consumer failed to consume with : %+v", string(debug.Stack()))
	}
}

func invokeNotificationHandler(notification string, cfg Config) {
	defer func() {
		if r := recover(); r != nil {
			invokeErrorHandler(
				fmt.Errorf("invokeNotificationHandler.panic: While processing %v, trace : %s", r, string(debug.Stack())),
				nil,
				cfg,
			)
		}
	}()
	cfg.NotificationHandler(notification)
}

func (sc *saramaConsumer) Health() (Health, error) {
	health := newHealth(sc.cfg)
	health.CanCosume = sc.canConsume.Load()
	conf := sarama.NewConfig()
	client, err := sarama.NewClient(sc.cfg.Address, conf)
	if err != nil {
		health.CanCosume = false
		return *health, err
	}
	defer client.Close() //nolint

	sc.checkConnectionState(client, health)

	sc.populatePartition(client, health)
	return *health, err
}

func (sc *saramaConsumer) populatePartition(client sarama.Client, health *Health) {
	for _, topic := range sc.cfg.Topics {
		partitions, err := client.Partitions(topic)
		if err != nil {
			Logger().Info(sc.cfg.TransactionID, "Error while getting partitions for topic: %s, %s", topic, err)
		} else {
			health.PerTopicPartitions[topic] = partitions
		}
	}
}

func (sc *saramaConsumer) checkConnectionState(client sarama.Client, health *Health) {
	coordinator, err := client.Coordinator(sc.cfg.Group)
	if err != nil {
		Logger().Debug(sc.cfg.TransactionID, "Failed to get Coordinator, error: %+v", err)
		health.CanCosume = false
		return
	}
	connected, err := coordinator.Connected()
	if err != nil {
		Logger().Debug(sc.cfg.TransactionID, "Failed to get Coordinator, error: %+v", err)
		health.CanCosume = false
	}

	if connected {
		for _, address := range sc.cfg.Address {
			health.ConnectionState[address] = true
		}
	}
}
