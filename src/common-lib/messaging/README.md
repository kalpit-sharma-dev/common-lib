# Messaging Lib

Messaging is a wrapper on top of [Confluent Kafka client](https://github.com/confluentinc/confluent-kafka-go) 
that provides additional tools for processing messages such as ordering.

This package is the next iteration of our previous 
[Platform Infrastructure Library Kafka implementation](https://gitlab.kksharmadevdev.com/platform/Platform-Infrastructure-lib/-/tree/master/messaging).

The previous library should be considered deprecated, with new use cases using
this library instead. For upgrading to use this library from the previous,
please see the section below.

> Note: Usage of anything in `messaging` via `messaging.go` and `messagingImpl.go` 
> instead of `messaging/consumer` or `messaging/producer` is deprecated. Use the new
> sub-packages instead.

## Third-Party Libraries

### [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go)

**License** [Apache License 2.0](https://github.com/confluentinc/confluent-kafka-go/blob/master/LICENSE)

`confluent-kafka-go` is Confluent's Golang client for Apache Kafka.

All Confluent Kafka client libraries utilize [librdkafka](https://github.com/edenhill/librdkafka)
as the underlying utility for communicating to Kafka.

In general, this library is bundled with `confluent-kafka-go`, however
if there are issues in your environment with loading this libarary, refer
to [the documentation for installing it](https://github.com/edenhill/librdkafka).

> Warning: This library is not supported on Windows. 
> For Windows development, use Docker or WSL2.

### [workerpool](https://github.com/gammazero/workerpool)
  
**License** [MIT License](https://github.com/gammazero/workerpool/blob/master/LICENSE)

`workerpool` provides helpful utilities for managing concurrent goroutine pools.

### Glide Dependencies

If using Glide for dependency management, the following will need
to be added for pulling the correct version of these 3rd party
libraries:

```yaml
- package: github.com/confluentinc/confluent-kafka-go
  version: v1.4.2
- package: github.com/gammazero/workerpool
  version: 1c428e5ee8aeada5fedba4e7af3c7cd61d57d1b9
```

## Kafka Consumer

The goal of the `consumer/` package is to provide throttling at the consumer level
and to enhance reliability for Kafka queue management.

Throttling will ensure the number of goroutines spawned never reaches above 
a pre-defined value (`SubscriberPerCore` \* Number of core) on the consumer's 
running machine.

Reliability is achieved and enhanced by `CommitMode: OnMessageCompletion` 
(described below), which will ensure that only processed messages will be 
committed to the Kafka cluster. In case of any panic from the consumer end, 
there will be no message loss and uncommitted messages can be processed on 
next run of the consumer.

This consumer needs to be instantiated once and takes parameters as described at
[the consumer README](consumer/README.md)

## Kafka Producer

The Kafka producer provides a means to publish messages to the Kafka stream
with various connection settings, and reconnect in case there is an issue.

The producer takes parameters as described in [the producer README](producer/README.md)

## Running a Kafka broker locally

For local testing and development, Docker Compose files have been provided for
running a Kafka broker. The included producer and consumer `example/` implementations
in each package can be run when connected to this compose file.

```bash
# start zookeeper and kafka
docker-compose -f docker-compose-kafka.yml up
```

### Seeing it all in action (example apps)

You can launch the consumer and producer samples to see them both work together.

> NOTE: consumer and producer in example have a 30 second wait at startup

1. Create a `.env` from `.env.sample` with your GitLab username and personal API
   token
2. Run the following:

```bash
# start zookeeper, kafka, consumer, and timed producer
docker-compose up
```

## Upgrading from Platform-Infrastructure-Lib

If you were previously using the Platform-Infrastructure-Lib's `messaging` implementation,
upgrade to this library was designed to require minimal changes. 

The interfaces for `consumer` and `publisher` are largely the same. Please refer
to [CHANGES.md](CHANGES.md) to see what has been added and removed and what applies
to your use case.
