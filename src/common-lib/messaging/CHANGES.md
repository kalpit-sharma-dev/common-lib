# Changes from platform-infrastructure-lib

The `messaging` package is a wrapper on top of [Confluent Kafka client](https://github.com/confluentinc/confluent-kafka-go) 
that provides additional tools for processing messages such as ordering.

This package is the next iteration of our previous 
[Platform Infrastructure Library Kafka implementation](https://gitlab.kksharmadevdev.com/platform/Platform-Infrastructure-lib/-/tree/master/messaging).

The previous library should be considered deprecated, with new use cases using
this library instead. 

If you were previously using the Platform-Infrastructure-Lib's `messaging` implementation,
upgrade to this library was designed to require minimal changes. The interfaces for `consumer` and `publisher` are largely the same. Refer to the notes below to see what has been added and removed and what applies
to your use case.

## Migration Status

You can get an idea of what packages are using which iteration of the Kafka common lib client and their status moving to this client [here](https://confluence.kksharmadevdev.com/x/sa-yBw).

## Consumer Config

- `CommitMode.PullOrderedWithOffsetReplay` removed (deprecated)
- `config.Retention` removed (it is a broker setting)
- `config.RebalanceTimeout` removed (was not being used)
- `config.HandleCustomOffsetStash` removed (deprecated)
- `config.MetadataFull` removed (not supported in librdkafka)
- `config.MaxQueueSize` added (local message queue max size before sending to workers)
- `config.EmptyQueueWaitTime` added (to avoid CPU skipes)
- `config.CommitIntervalMs` added (interval for committing stored offsets)
- fix: for `ConsumerMode: PullOrdered`, if an error is returned by `config.MessageHandler`
  after all retries from `config.RetryCount` are exhausted, the message and error will
  be sent to `config.ErrorHandler` for the consumer to react to the fail

## Consumer Service interface

- `c.CloseWait()` added (wait for queued messages to process)

### How the consumer works behind the scenes

Messages are now polled on a background thread, then put into a local queue to distribute
the work. This means that a constant heartbeat will be sent, so the consumer will
not run  nto a session timeout issue from taking too long to process a message batch
before getting new events again, reducing rebalances in the consumer group.

For more info: https://docs.confluent.io/platform/current/clients/consumer.html#message-handling

## Producer Package

- renamed package from `publisher` to `producer` for consistency

## Producer Config

- `CompressionType` changed to string
- `config.ReconnectIntervalInSecond` removed (no longer needed)
- `config.MaxReconnectRetry` removed (no longer needed)
- `config.CleanupTimeInSecond` removed (no longer needed)
- `config.CircuitBreaker` removed (no longer needed) (replaced by settings below)
- `config.CommandName` removed (no longer needed)
- `config.CircuitProneErrors` removed (no longer needed)
- `config.ReconnectBackoffMs` added
- `config.ReconnectBackoffMaxMs` added
- `config.MessageTimeoutMs` added (max time before sending message to broker successfully)
- `config.MessageSendMaxRetries` added
- `config.RetryBackoffMs` added
- `config.QueueBufferingMaxMessages` added
- `config.QueueBufferingMaxKbytes` added
- `config.QueueBufferingMaxMs` added
- `config.EnableIdempotence` added (ensure original produce order on retry)
- `config.ProduceChannelSize` added

## Sync Producer

- `p.ProduceWithReport()` added (returns delivery report)
- `p.Close()` added
- `p.Health()` added (for consitency with consumer Health())
- `Health()` and `status.Status()` removed (in favor of p.Health())
- `p.Publish()` renamed to `p.Produce()`

## Async Producer (added)

- `p.ProduceChannel()` added
- `p.DeliveryReportChannel()` added
- `p.Flush()` added
- `p.Close()` added
- `p.Health()` added

## Message

- `m.Key` changed from `Encoder` to `[]byte`
- `m.Value` changed from `Encoder` to `[]byte`

## Health

- `PerTopicPartitions` removed (cannot offer)
- `CanCosume` removed (can always consume (background process))
