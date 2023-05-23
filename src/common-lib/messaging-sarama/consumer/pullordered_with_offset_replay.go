// Package consumer - This is a specific consumer mode implementation with the cappability to replay offset,
// based on the set of offset for the topic which is received as a part of the method handler
// Config.HandleCustomOffsetStash(), this method will be implemented by specific application.
// The integer code that is associated with this mode is -3, so the config.ConsumerMode needs to be set to -3
// for this mode to get initiated.
// NOTE:
// Replay feature is essentially based on the kafka metadata handled by the specific application
// and doesn't read from any specific offset from kafka.The application should return
// the offset which they want to replay during consumer startup.
package consumer

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	cluster "github.com/bsm/sarama-cluster"
)

var errorMap = make(map[string]string)

const (
	// Kafka101 - Error code for the Handling custom offset failure
	Kafka101 = "Kafka101"

	// Kafka102 - processing custom offset failure
	Kafka102 = "Kafka102"

	partitionKeyFormat = "%s_%d"
)

func init() {
	errorMap[Kafka101] = "%s : Error Handling custom offset: Reason : %+v"
	errorMap[Kafka102] = "%s : Error processing custom offset %v for Partition: %v and Topic: %v by MessageHandler: Reason: %v"
}

//Returns a new instance for the pullOrderedWithOffsetReplay consumer mode
func getNewpullOrderedWithOffsetReplay(conf Config) *pullOrderedWithOffsetReplay {
	porcm := &pullOrderedWithOffsetReplay{
		cfg:               conf,
		partitionKeyStore: make(map[string]bool),
	}

	return porcm
}

//pullOrderedWithOffsetReplay - Stratgy Handler
type pullOrderedWithOffsetReplay struct {
	cfg Config
	//This store is required to identitfy if the current partition is already being processed
	//by the consumer or not in case of rebalance and reprocessing.
	partitionKeyStore      map[string]bool
	partitionKeyStoreMutex sync.Mutex
}

//pull: gets called from sarama consumer instance for start pulling messages.
//Method call hierarchy for the pull function to avoid confusion:
//			pull --> consume --> processMessage
//nolint:gosimple
func (owr *pullOrderedWithOffsetReplay) pull(sc *saramaConsumer) {
	log := Logger() // nolint - false positive
	log.Trace(owr.cfg.TransactionID, "Starting Consumer in pullOrderedWithOffsetReplay mode \n")
	partitions := make(chan cluster.PartitionConsumer, 1)

	go owr.consume(sc, partitions)

	// consume partitions all the partitions current consumer is listening to goes to the channel.
	//The consumer of the partition channel listens to it and starts processing the messages one by one.
	for {
		select {
		case part, ok := <-sc.consumer.Partitions():
			if !ok {
				return
			}
			log.Trace(owr.cfg.TransactionID, "Sending partition for Consumption : %v\n", part.Partition())
			partitions <- part
		}
	}
}

// start a separate goroutine for each partition the consumer is listening to for a topic
func (owr *pullOrderedWithOffsetReplay) consume(sc *saramaConsumer, partitions <-chan cluster.PartitionConsumer) {
	for {
		pc := <-partitions
		go owr.processPartition(sc, pc)
	}
}

//processPartition: this method is called in seperate goroutene as this method is for processing messages for a specific partition.
func (owr *pullOrderedWithOffsetReplay) processPartition(sc *saramaConsumer, pc cluster.PartitionConsumer) {
	log := Logger() // nolint - false positive
	log.Debug(owr.cfg.TransactionID, "Starting Consumer Routine to Consume from partition : %v\n", pc.Partition())

	//If the partition is a net new one and not present in the local cache fetch the application offset and reset to the corresponding
	//offset and then start processing the message.
	if !owr.isPartitionAlreadyConsumed(pc) {
		log.Debug(owr.cfg.TransactionID, "Start consuming a new Partition %v for topic %v\n", pc.Partition(), pc.Topic())
		err := owr.handleOffsetforPartition(sc, pc)
		//If error occurs while fetching application offset we return and stop consumption for the partion in this mode.
		if err != nil {
			//This error should not stop normal message consumption. Kafka101: Error in handling custom offset
			log.Error(owr.cfg.TransactionID, Kafka101, errorMap[Kafka101], "processPartition", err.Error())
		}
	}
	//Note: In case of rebalance if the this partition is removed from current consumer , then it will close the message channel for that partition
	//in which case the goroutene handling the current partition will not be hung on the message channel and it will be freed.
	for msg := range pc.Messages() {
		sc.canConsume.Store(true)
		consumerMessage := newMessage(msg)
		sc.commitStrategy.onPull(consumerMessage.GetTransactionID(), msg.Topic, msg.Partition, msg.Offset)
		owr.processMessage(consumerMessage, sc)
	}

}

//isPartitionAlreadyConsumed: checks if the current partition is already being consumed , by checking in the local cache
//If the current partition is not present in the partitionKeyStore then add it in the store with value as the PartitonConsumer instance.
func (owr *pullOrderedWithOffsetReplay) isPartitionAlreadyConsumed(pc cluster.PartitionConsumer) bool {
	partitionKeyStr := owr.partitionKeyStore

	//The partition key store has key as topicName_partitionId because the same consumer might be listening to multiple Topics
	partitionKey := owr.getPartitionKeyFrm(pc)

	//If partition key is present in the store ,then that means the partition for the topic
	//is already being listened on by the current consumer.
	//Below needs to be accessed in lock as t might be in the process of getting updated due to Rebalance.
	owr.partitionKeyStoreMutex.Lock()
	defer owr.partitionKeyStoreMutex.Unlock()
	isPartitionKeyConsumed := partitionKeyStr[partitionKey]

	//The partition will be consumed the first time
	partitionKeyStr[partitionKey] = true
	return isPartitionKeyConsumed
}

//Gets the custom offsetstash store implemented and passed by the application in the config.
//Gets the offsetStashList for the current topic.
func (owr *pullOrderedWithOffsetReplay) handleOffsetforPartition(sc *saramaConsumer, pc cluster.PartitionConsumer) error {
	defer func() {
		if r := recover(); r != nil {
			invokeErrorHandler(
				fmt.Errorf("handleOffsetforPartition.Panic: While processing %v, trace : %s", r, string(debug.Stack())),
				&Message{},
				sc.cfg,
			)
		}
	}()
	var err error
	var retryCount int64
	var customOffsetStashLst []OffsetStash
	//This will be a map of key= topic and value= array of OffsetStash
	if owr.cfg.HandleCustomOffsetStash != nil {
		for retryCount = 0; retryCount < sc.cfg.RetryCount; retryCount++ {

			customOffsetStashLst, err = owr.cfg.HandleCustomOffsetStash(pc.Topic(), pc.Partition())
			if err == nil {
				break
			}
			time.Sleep(sc.cfg.RetryDelay)
		}
	}

	if err != nil {
		return fmt.Errorf("Custom Offset Handling Failed with error: %+v", err)
	}

	if customOffsetStashLst != nil {
		owr.processOffsetStash(customOffsetStashLst, sc, pc)

	}
	return err
}

//1. Get offsetStash for the current pc
//2. Generate Message with the offsetstash data
//3. Call the message handler from the config with the above generated message
//returns Error
func (owr *pullOrderedWithOffsetReplay) processOffsetStash(offsetStashLst []OffsetStash, sc *saramaConsumer, pc cluster.PartitionConsumer) {
	log := Logger()
	var err error
	var retryCount int64
	for _, offsetStash := range offsetStashLst {
		//Retry in case of message handler failure
		for retryCount = 0; retryCount < sc.cfg.RetryCount; retryCount++ {
			msg := newMessageFrmOffsetStash(offsetStash)
			err = invokeMessageHandler(msg, sc.cfg)
			if err == nil {
				break
			}

			time.Sleep(sc.cfg.RetryDelay)
		}
		if err != nil {
			//Kafka102: Error occured when consumer was processing the message through MessageHandler
			log.Error(owr.cfg.TransactionID, Kafka102, errorMap[Kafka102], "pullOrderedWithOffsetReplay", offsetStash.Offset, offsetStash.Partition, offsetStash.Topic, err)
		}

	}

}

//Generate consumer.Message from the offsetStash
func newMessageFrmOffsetStash(offsetStash OffsetStash) Message {
	consumerMessage := Message{
		Message:           offsetStash.Message,
		Offset:            offsetStash.Offset,
		Partition:         offsetStash.Partition,
		Topic:             offsetStash.Topic,
		PulledDateTimeUTC: offsetStash.PulledDateTimeUTC,
		headers:           offsetStash.Header,
		transactionID:     offsetStash.TransactionID,
	}

	return consumerMessage
}

//refreshPartitionKeyStore: will be called from consumer.consumeRebalanceNotification() function
//based on whether the consumer mode is pullOrderedWithOffsetReplay only
//rebalanceNotification: the notification object when rebalance starts, ends with error or ends successfully
func (owr *pullOrderedWithOffsetReplay) refreshPartitionKeyStore(rebalanceNotification *cluster.Notification) {
	log := Logger() // nolint - false positive
	partitionKeyStr := owr.partitionKeyStore

	//Based on the type of rebalance notification corrective action is taken
	switch rebalanceNotification.Type {
	case cluster.RebalanceStart:
		log.Debug(owr.cfg.TransactionID, "Rebalance Started ...")
	case cluster.RebalanceError:
		log.Debug(owr.cfg.TransactionID, "Rebalance Error occurred... %v", rebalanceNotification)
	case cluster.RebalanceOK:
		log.Debug(owr.cfg.TransactionID, "Rebalance Success ... New partiton set: %v", rebalanceNotification.Current)

		//Iterate over the released topic/partiton map and remove the
		//partition key from the partitionKeyStore which are released during rebalance
		owr.partitionKeyStoreMutex.Lock()
		defer owr.partitionKeyStoreMutex.Unlock()
		for topic, partitionLst := range rebalanceNotification.Released {
			for partition := range partitionLst {
				partitionKey := fmt.Sprintf(partitionKeyFormat, topic, partition)
				delete(partitionKeyStr, partitionKey)
			}
		}

	}
}

func (owr *pullOrderedWithOffsetReplay) getPartitionKeyFrm(pc cluster.PartitionConsumer) string {
	topic := pc.Topic()
	partition := pc.Partition()
	//The partition key store has key as topicName_partitionId because the same consumer might be listening to multiple Topics
	partitionKey := fmt.Sprintf(partitionKeyFormat, topic, partition)
	return partitionKey
}

func (owr *pullOrderedWithOffsetReplay) processMessage(consumerMessage Message, sc *saramaConsumer) {
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
