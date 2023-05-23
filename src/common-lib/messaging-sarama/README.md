# DEPRECATION NOTICE

`messaging-sarama` is deprecated, use `messaging` instead for newer services.

# Messaging Services

messaging-sarama is a wrapper on top of [Sarama](https://github.com/Shopify/sarama) wrapper → A Kafka client library

## Import paths

```
"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging-sarama/consumer"
"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/messaging-sarama/publisher"
```

### Third-Party Libraties

- [Sarama](https://github.com/Shopify/sarama)
- **License** [MIT License](https://github.com/Shopify/sarama/blob/master/LICENSE)
- **Description**
  - Sarama is an MIT-licensed Go client library for Apache Kafka version 0.8 (and later).
- [Sarama Cluster](https://github.com/bsm/sarama-cluster)
- **License** [MIT License](https://github.com/bsm/sarama-cluster/blob/master/LICENSE)
- **Description**
  - Cluster extensions for Sarama, the Go client library for Apache Kafka 0.9

## Consumer

The goal of this kafka consumer is to provide throttling at consumer level and to enhance reliability for kafka queue management.

Throttling will ensure the number of GO routines spawned never reaches above pre-defined value (SubscriberPerCore \* Number of core) on consumer running machine.

Reliability is achieved and(or) enhanced by OnMessageCompletion CommitMode (described below) which will ensure that only processed messages will be committed to Kafka cluster.
In case of any panic from consumer end, there will be no message loss and uncommitted messages can be processed on next run of consumer.

This consumer needs to be instantiated once and takes parameters as described at [README](consumer/README.md)

Features available with this Consumer :

- Pull - This will call Sarama's Pull method and will fetch messages from Kafka Broker
- MarkOffset - This will commit the offset for a given partition and topic
- Close - This will close consumer connection with Kafka cluster
- Health - this will provide Kafka consumer health status

This consumer also provides notification handler which can get notifications when re-balancing in kafka system happens.

This consumer uses [Worker Pool Library](https://github.com/goinggo/work) to implement throtting for Kafka messages.

Any Application or Microservices are expected to handle below scenario in order to use this library efficiently :

- Message duplicate handling - Since message's offset are committed only when it is processed, this is possible that in case of any panic (application box goes down or if it needs a re-start) , uncommitted messages will be read again and hence duplicate messages may arrive. Application or services using this consumer should handle this duplicate scenario

- Error handling - A consumer implements a timeout of 60 seconds (default) for every message processing. In case of any panic or error (system or user error which can be re-triable or non-retriable) , error handling is expected to be done efficiently by application (or service). If no error handling is done, consumer treats the message as processed after time-out. Corresponding GO routines still exist in the worker pool but they will go under natural death after their life-cycles.

This may also lead to complete blocking worker pool (if no of unprocessed messages == size of worker pool) though they would be marked as completed by this Consumer.
Reaching to that case will block further incoming messages from Kafka brokers.

## Publisher

Publisher is a wrapper on top of [Sarama](https://github.com/Shopify/sarama) → A Kafka client library

- [MIT License](https://github.com/Shopify/sarama/blob/master/LICENSE)

**Additional details on top of Sarama**

This package is providing below additional features on top of Sarama

- Reconnect to Kafka
  - Reconnect to another Kafka instance as soon as any one node goes down or stopped responding
- Publish a message using Context
  - Publish a message using to enable timeout for an messaes instead of waiting for longer to avoid increase in memory consumption

**Limitation**

- Today only support Sync producer, Async need to implementation in case needed for any MS

[README](publisher/README.md)
