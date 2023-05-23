# DEPRECATION NOTICE

`messaging-sarama/consumer` is deprecated, consider using `messaging/consumer` instead

# Consumer

Consumer is a wrapper on top of [Sarama](https://github.com/Shopify/sarama) → A Kafka client library

## Additional details on top of Sarama

This package is providing below additional features on top of Sarama

- Limits subscriber routines based on input
  - To help the consumer to keep consuming messages based on a threshold and saves from sudden burst of messages
  - This threshold can be tweacked and can be set according to the hardware capability
- Ability to mark offset when the consumer has done processing of the respective message
  - If for some reason, the message did not process, it will be available in Kafka Topic for re-processing
- Ability to process messages sequentially from Kafka
  - This reads messages from multiple partitions to provide parallel processing
- Health(Lag) of the topic based on the consumer group

## Configuration

Use `consumer.NewConfig()` to populate the default values

```go
// Config is a struct used by Kafka Consumer
type Config struct {
	//Address : is a Kafka Broker IPs
	Address []string

	//Group : is  a Consumer Group name
	Group string

	//Topics to be consumed
	Topics []string

	//SubscriberPerCore : Number of maximum go routines for processing per Core
	SubscriberPerCore int

	//CommitMode : Specify which commit mode you want to use
	// OnPull means offset will be committed on Pull before message processing starts
	// OnMessageCompletion means offset will be committed on after message processing completes
	CommitMode commitMode

	// ConsumerMode : Specify which consumer mode you want to use
	// PullUnOrdered : Messages will be processed in batches
	// PullOrdered : This Mode claims that message will not lost under any system failure.
	// NOTE: This needs more partition as messages are processed sequentially / partition.
	// Default is PullUnOrdered
	ConsumerMode consumerMode

	//Timeout After this duration message offset will be committed without releasing worker
	Timeout time.Duration

	// The retention duration for committed offsets.
	// If zero, disabled (in which case the `offsets.retention.minutes` option on the
	// broker will be used).  Kafka only supports precision up to milliseconds;
	// nanoseconds will be truncated. Requires Kafka
	// broker version 0.9.0 or later. (0 is : disabled).
	// default : 1 Minute
	Retention time.Duration

	// RebalanceTimeout: The maximum allowed time for each worker to join the group once a rebalance has begun.
	// This is basically a limit on the amount of time needed for all tasks to flush any pending
	// data and commit offsets. If the timeout is exceeded, then the worker will be removed from
	// the group, which will cause offset commit failures (default 60s).
	RebalanceTimeout time.Duration

	//This will ge used as a callback function for message handling
	MessageHandler func(Message) error

	//This function is only used in case of a specific consumer-mode for now.
	//The consumer mode with which this function is used is PullOrderedWithOffsetReplay
	//The function will return a map whose
	//key:  Will be the Topicname string
	//value: Will be array of type OffsetStash
	//The implementation of this function should ideally include the logic to fetch the offset for the partitions
	//and return the same as the map output defined for the function signature.
	HandleCustomOffsetStash func(Config) (map[string][]OffsetStash, error)

	// If available, any errors that occurred while consuming are returned on
	// the Errors Handler (default disabled).
	ErrorHandler func(error, *Message)

	// If available, rebalance notification will be returned on the
	// Notifications Handler (default disabled).
	NotificationHandler func(string)

	// The initial offset to use if no offset was previously committed.
	// Should be OffsetNewest or OffsetOldest. Defaults to OffsetNewest.
	OffsetsInitial int64

	// Whether to maintain a full set of metadata for all topics, or just
	// the minimal set that has been necessary so far. The full set is simpler
	// and usually more convenient, but can take up a substantial amount of
	// memory if you have many topics and partitions. Defaults to true.
	MetadataFull bool

	// RetryDelay: This flag is used in case you want to wait for some time before retrying the the failed message.
	// Defaults to 30s.
	// NOTE: This only Works with PullOrdered Consumer Mode
	RetryDelay time.Duration

	// RetryCount : Specify how meany times a message should be retried before dropping the message
	// Defaults to 10.
	// NOTE: This only Works with PullOrdered Consumer Mode
	RetryCount int64
}
```

## Example

Simple example having message handler

```go
package main

import (
    "fmt"

    "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging-sarama/consumer"
)

func main() {
    cfg := consumer.NewConfig()
    cfg.Address = []string{"localhost:9092"}
    cfg.Topics = []string{"test"}
    cfg.Group = "simpleConsumer"
    cfg.MessageHandler = messageHandler
    srvc, err := consumer.New(cfg)
    if err != nil {
        fmt.Println(err)
        return
    }
    srvc.Pull()
}

func messageHandler(msg consumer.Message) error {
    fmt.Printf("%+v\n", msg)
    fmt.Printf("%s\n", string(msg.Message))
    return nil
}
```

- [Complete example for ParallelConsumer](https://gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/infra-lib/tree/master/messaging/consumer/example/parallel/consumer.go)

- [Complete example for SequentialConsumer](https://gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/infra-lib/tree/master/messaging/consumer/example/sequential/consumer.go)
