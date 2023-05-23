# Message Consumer

Consumer is a wrapper on top of [Confluent Kafka client](https://github.com/confluentinc/confluent-kafka-go) 
that provides additional tools for processing messages such as ordering.

## Additional details

This package is providing with the following additional features on top of Confluent:

- Limits subscriber routines based on input
  - To help the consumer to keep consuming messages based on a threshold and 
    saves from a sudden burst of messages
  - This threshold can be tweaked and can be set according to the hardware capability
    via `SubscriberPerCore`
- Ability to mark the offset when a consumer has completed processing 
  of the respective message
  - If for some reason, the message did not process, it will still be available in
    the Kafka Topic for re-processing
  - Configured with the `CommitMode: OnMessageCompletion` option
- Ability to process messages sequentially from Kafka
  - This reads messages from multiple partitions to provide parallel processing
  - Configured with the `ConsumerMode: PullOrdered` option
- Health of the topic based on the consumer group

Interface available with this Consumer :

- `Pull` - This will fetch messages from the Kafka Broker.
- `MarkOffset` - This will commit the offset for a given partition and topic.
- `Close` - This will close the consumer connection with the Kafka cluster.
- `Health` - This will provide Kafka consumer connection status.

This consumer also provides a notification handler which can get notifications 
when re-balancing in the Kafka system happens.

This consumer uses the worker pool to implement
throttling for Kafka messages.

Consumers of this library are expected to handle the below scenarios in order 
to use this library effectively:

- **Message duplicate handling** - For `OnMessageCompletion`, since the message 
  offsets are committed only when the message has been processed, it is possible
  that in case of any panic (for example, the application box goes down or if it
  needs a restart), any uncommitted messages will be read again. This means that
  duplicate messages may arrive. Application or services using this library should
  handle this duplicate scenario.

- **Error handling** - A Kafka consumer implements a timeout of 60 seconds by default
  for every message processed. In case of any panic or error (a system or user
  error which can be retriable or non-retriable), error handling is expected to 
  be done by the consuming application (or service). If no error handling is done,
  the library treats the message as processed after the timeout. Corresponding goroutines
  will still exist in the worker pool, but they will go under natural death
  after their life-cycles.

This may also lead to blocking the worker pool (if the number of 
unprocessed messages == size of worker pool) though they would be marked as 
completed. Reaching to that case will block further incoming messages from the 
Kafka brokers.

## Configuration

Use `consumer.NewConfig()` to populate the default values

```go

// Config is a struct used by Kafka Consumer
type Config struct {
	// Address: Kafka Broker IPs
	Address []string

	// Group: A Consumer Group name
	Group string

	// Topics: Topics to be consumed
	Topics []string

	// SubscriberPerCore: Number of maximum go routines for processing per Core
	SubscriberPerCore int

	// CommitMode: Specify which commit mode you want to use
	// OnPull means offset will be committed on Pull before message processing starts
	// OnMessageCompletion means offset will be committed on after message processing completes
	// Default: OnPull
	CommitMode commitMode

	// ConsumerMode: Specify which consumer mode you want to use
	// PullUnOrdered : Messages will be processed in batches
	// PullOrdered : This Mode claims that message will not lost under any system failure.
	//
	// NOTE: This needs more partition as messages are processed sequentially / partition.
	// Default: PullUnOrdered
	ConsumerMode consumerMode

	// Timeout: After this duration message offset will be committed without releasing worker
	Timeout time.Duration

	// MessageHandler: This will get used as a callback function for message handling (must be thread safe)
	MessageHandler func(Message) error

	// ErrorHandler: If available, any errors that occurred while consuming are returned on this handler (must be thread safe)
	ErrorHandler func(error, *Message)

	// NotificationHandler: If available, rebalance notification will be returned on this handler
	NotificationHandler func(string)

	// OffsetsInitial: The initial offset to use if no offset was previously committed.
	// Should be OffsetNewest or OffsetOldest.
	// Default: OffsetNewest.
	OffsetsInitial int64

	// RetryDelay: This flag is used in case you want to wait for some time before retrying the the failed message.
	// Default: 30s
	RetryDelay time.Duration

	// RetryCount: Specify how meany times a message should be retried before dropping the message
	// Default: 10
	RetryCount int64

	//TransactionID: Default transaction id to be used by the consumer
	TransactionID string

	// Partitions: Max Number of partitions to be consumed by consumer
	// NOTE: This number should always be bigger than the number of partitions available in the production for all the topics
	// Default: 500.
	Partitions int

	// MaxQueueSize: Max messages to keep in local queue
	// NOTE: the greater this is, the more memory the service uses processing messages does not keep up
	// Default: 100
	MaxQueueSize int

	// EmptyQueueWaitTime: specifies the time to wait when we encounter an empty queue of messages
	// NOTE: this is to keep CPU from spiking at the cost of locking the main process thread, keep it low
	// Default: 500ms
	EmptyQueueWaitTime time.Duration

	// CommitIntervalMs: interval between committing stored offsets
	// Default: 5000
	CommitIntervalMs int
}
```

## Example

Simple example for having a message handler:

```go
package main

import (
    "fmt"

    "gitlab.kksharmadevdev.com/platform/Platform-Infrastructure-lib/messaging/consumer"
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

[A runnable example is available](examples/).


## Consumer Performance Testing
To test the performance of kafka consumer use test in the `integration_test.go` file.

To run this test:
- you need to have kafka up and running on `localhost:9092`
- run the tests with `integration` flag: `go test -tags=integration` 

As for the results, you're interested in the logs like:
```text
topic deleted
topic created
published 5000 messages
processed 5000 messages, in 141.303217ms
topic deleted
```
On my machine (Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz 4 cores/8 threads + 16 Gb RAM) 5000 messages spread across 10 partitions are processed in 150ms on average. 
