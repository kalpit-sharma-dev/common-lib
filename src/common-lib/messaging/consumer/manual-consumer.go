package consumer

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

//go:generate mockgen -package consumer -source=manual-consumer.go -destination=manual-consumer_mock.go .

// ManualService interface.
type ManualService interface {
	// Poll polls message from Kafka topic.
	// returns nil on timeout, otherwise message.
	Poll(timeout int) *kafka.Message
	// CommitOffset commits offset for partition and topic.
	CommitOffset(topic string, partition int32, offset int64) error
	// Close consumer connection.
	Close() error
	// Health returns consumer's health.
	Health() (Health, error)
	// Pause consumption from the specified topic.
	// Use with caution in PullUnOrdered mode as your worker pool can be filled with messages from this partition and prevent consumption of others.
	Pause(topic string, partition int32, offset int64) error
	// Resume consumption from the specified topic.
	Resume(topic string, partition int32, offset int64) error
	// PauseAll Pauses consumption from all partitions assigned to this consumer.
	PauseAll() error
	// ResumeAll Resumes consumption from all partitions assigned to this consumer.
	ResumeAll() error
}

// ManualKafkaConsumer consumer of kafka messages.
type ManualKafkaConsumer struct {
	config   *Config
	consumer consumer

	healthMx *sync.RWMutex
	health   *Health

	errorLog ConsumerLogger
	infoLog  ConsumerLogger
	debugLog ConsumerLogger

	pausedPartitionsMx *sync.Mutex // it's rarely used and always used to mutate the state so there is no point in using RWMutex
	pausedPartitions   map[partition]kafka.TopicPartition
}

// NewManualConsumer creates a new ManualConsumer instance with config settings.
func NewManualConsumer(config *Config) (*ManualKafkaConsumer, error) {
	if len(config.Address) == 0 || config.Group == "" || len(config.Topics) == 0 {
		return nil, fmt.Errorf("some of the values are blank. Address : %v, Group : %s and Topics : %v",
			config.Address, config.Group, config.Topics)
	}
	configMap := kafka.ConfigMap{
		"bootstrap.servers":               strings.Join(config.Address, ","),
		"group.id":                        config.Group,
		"go.application.rebalance.enable": true,
		"auto.offset.reset":               getResetPosition(config),
		"enable.auto.commit":              false, // for this Consumer, manual commit is the only way to commit offsets!
		"enable.auto.offset.store":        false,
	}
	if config.EnableKafkaLogs {
		configMap["go.logs.channel.enable"] = true
		configMap["log_level"] = config.LogLevel
		if len(config.LoggedComponents) > 0 {
			configMap["debug"] = config.LoggedComponents
		}
	}

	c, err := kafka.NewConsumer(&configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %s", err)
	}

	kc := &ManualKafkaConsumer{
		config:   config,
		consumer: c,

		healthMx: &sync.RWMutex{},
		health:   newHealth(),

		errorLog: config.ErrorLog,
		infoLog:  config.InfoLog,
		debugLog: config.DebugLog,

		pausedPartitionsMx: &sync.Mutex{},
		pausedPartitions:   make(map[partition]kafka.TopicPartition),
	}
	func() {
		kc.healthMx.Lock()
		defer kc.healthMx.Unlock()
		kc.health.Group = config.Group
		kc.health.Topics = config.Topics
		kc.health.Address = config.Address
	}()

	if config.EnableKafkaLogs {
		if ch := c.Logs(); ch != nil {
			go kc.logKafkaLogs(ch) // logs chan is closed by kafka on Close()
		}
	}

	rebalanceCb := func(cons *kafka.Consumer, ev kafka.Event) error {
		switch e := ev.(type) {
		case kafka.AssignedPartitions:
			kc.invokeNotificationHandler(fmt.Sprintf("%+v", e))
			kc.processAssignedPartitions(e)
			return nil
		case kafka.RevokedPartitions:
			kc.invokeNotificationHandler(fmt.Sprintf("%+v", e))
			kc.processRevokedPartitions()
			return nil
		default:
			return nil
		}
	}

	err = kc.consumer.SubscribeTopics(kc.config.Topics, rebalanceCb)
	if err != nil {
		return nil, err
	}
	return kc, nil
}

func (c *ManualKafkaConsumer) processAssignedPartitions(p kafka.AssignedPartitions) {
	if err := c.ResumeAll(); err != nil {
		c.invokeErrorHandler(err)
	}

	if err := c.consumer.Assign(p.Partitions); err != nil {
		c.invokeErrorHandler(err)
		return
	}
	c.healthMx.Lock()
	defer c.healthMx.Unlock()
	c.health.ConnectionState = true
}

func (c *ManualKafkaConsumer) processRevokedPartitions() {
	if err := c.ResumeAll(); err != nil {
		c.invokeErrorHandler(err)
	}
	if err := c.consumer.Unassign(); err != nil {
		c.invokeErrorHandler(err)
	}
}

// CommitOffset commits offset for partition.
func (c *ManualKafkaConsumer) CommitOffset(topic string, partition int32, offset int64) error {
	// consumer starts at N, so we have to store N + 1
	// https://docs.confluent.io/5.5.0/clients/librdkafka/rdkafka_8h.html#ab96539928328f14c3c9177ea0c896c87
	_, err := c.consumer.CommitOffsets([]kafka.TopicPartition{{
		Topic:     &topic,
		Partition: partition,
		Offset:    kafka.Offset(offset + 1),
	}})
	return err
}

// Poll polls a single kafka message.
// Returns nil on timeout or errors, message otherwise.
// If kafka return an error instead of message, it is handled by the error handler and method returns nil.
func (c *ManualKafkaConsumer) Poll(timeoutMs int) *kafka.Message {
	m := c.consumer.Poll(timeoutMs)

	switch e := m.(type) {
	case *kafka.Message:
		return e
	case kafka.Error:
		// Errors should generally be considered as informational, the client will try to automatically recover
		if e.Code() == kafka.ErrAllBrokersDown {
			func() {
				c.healthMx.Lock()
				defer c.healthMx.Unlock()
				c.health.ConnectionState = false
			}()
		}
		c.invokeErrorHandler(e)
	default:
		// nothing
	}
	return nil
}

// Close stops pulling messages from kafka.
func (c *ManualKafkaConsumer) Close() error {
	c.pausedPartitionsMx.Lock()
	defer c.pausedPartitionsMx.Unlock()
	for part := range c.pausedPartitions {
		delete(c.pausedPartitions, part)
	}
	return c.consumer.Close()
}

// Health returns consumer's health.
func (c *ManualKafkaConsumer) Health() (Health, error) {
	c.healthMx.RLock()
	defer c.healthMx.RUnlock()
	return *c.health, nil
}

// Pause consumption from the specified topic.
func (c *ManualKafkaConsumer) Pause(topic string, part int32, offset int64) error {
	tp := kafka.TopicPartition{
		Topic:     &topic,
		Partition: part,
		Offset:    kafka.Offset(offset),
	}
	p := partition{
		topic:     *tp.Topic,
		partition: tp.Partition,
	}
	c.pausedPartitionsMx.Lock()
	defer c.pausedPartitionsMx.Unlock()
	if _, ok := c.pausedPartitions[p]; ok {
		return nil // already paused
	}
	if err := c.consumer.Pause([]kafka.TopicPartition{tp}); err != nil {
		return err
	}
	c.pausedPartitions[p] = tp
	return nil
}

// Resume consumption from the specified topic.
func (c *ManualKafkaConsumer) Resume(topic string, part int32, offset int64) error {
	tp := kafka.TopicPartition{
		Topic:     &topic,
		Partition: part,
		Offset:    kafka.Offset(offset),
	}
	p := partition{
		topic:     *tp.Topic,
		partition: tp.Partition,
	}
	c.pausedPartitionsMx.Lock()
	defer c.pausedPartitionsMx.Unlock()
	ptp, ok := c.pausedPartitions[p]
	if !ok {
		return nil // not paused
	}
	if err := c.consumer.Resume([]kafka.TopicPartition{ptp}); err != nil {
		return err
	}
	delete(c.pausedPartitions, p)
	return nil
}

// PauseAll Pauses consumption from all partitions assigned to this consumer.
func (c *ManualKafkaConsumer) PauseAll() error {
	c.pausedPartitionsMx.Lock()
	defer c.pausedPartitionsMx.Unlock()
	parts, err := c.consumer.Assignment()
	if err != nil {
		return err
	}

	var toPause []kafka.TopicPartition
	for _, tp := range parts {
		p := partition{
			topic:     *tp.Topic,
			partition: tp.Partition,
		}
		if _, ok := c.pausedPartitions[p]; ok {
			continue // already paused
		}
		toPause = append(toPause, tp)
	}

	if err = c.consumer.Pause(toPause); err != nil {
		return err
	}
	for _, tp := range toPause {
		p := partition{
			topic:     *tp.Topic,
			partition: tp.Partition,
		}
		c.pausedPartitions[p] = tp
	}
	return nil
}

// ResumeAll Resumes consumption from all partitions assigned to this consumer.
func (c *ManualKafkaConsumer) ResumeAll() error {
	c.pausedPartitionsMx.Lock()
	defer c.pausedPartitionsMx.Unlock()
	var toResume []kafka.TopicPartition
	for _, tp := range c.pausedPartitions {
		toResume = append(toResume, tp)
	}
	if len(toResume) > 0 {
		if err := c.consumer.Resume(toResume); err != nil {
			return err
		}
	}
	for part := range c.pausedPartitions {
		delete(c.pausedPartitions, part)
	}
	return nil
}

func (c *ManualKafkaConsumer) logKafkaLogs(ch <-chan kafka.LogEvent) {
	for l := range ch {
		switch l.Level {
		case 0:
			c.errorf("[%s] [%s] %s", l.Name, l.Tag, l.Message)
		case 1:
			c.errorf("[%s] [%s] %s", l.Name, l.Tag, l.Message)
		case 2:
			c.errorf("[%s] [%s] %s", l.Name, l.Tag, l.Message)
		case 3:
			c.infof("[%s] [%s] %s", l.Name, l.Tag, l.Message)
		case 4:
			c.infof("[%s] [%s] %s", l.Name, l.Tag, l.Message)
		case 5:
			c.infof("[%s] [%s] %s", l.Name, l.Tag, l.Message)
		case 6:
			c.infof("[%s] [%s] %s", l.Name, l.Tag, l.Message)
		case 7:
			c.debugf("[%s] [%s] %s", l.Name, l.Tag, l.Message)
		}
	}
}

func (c *ManualKafkaConsumer) errorf(format string, v ...interface{}) {
	if c.errorLog != nil {
		c.errorLog.Printf(format, v...)
	}
}
func (c *ManualKafkaConsumer) infof(format string, v ...interface{}) {
	if c.infoLog != nil {
		c.infoLog.Printf(format, v...)
	}
}
func (c *ManualKafkaConsumer) debugf(format string, v ...interface{}) {
	if c.debugLog != nil {
		c.debugLog.Printf(format, v...)
	}
}

func (c *ManualKafkaConsumer) invokeNotificationHandler(notification string) {
	defer func() {
		if r := recover(); r != nil {
			c.invokeErrorHandler(
				fmt.Errorf("invokeNotificationHandler.panic: While processing %v, trace : %s", r, string(debug.Stack())),
			)
		}
	}()
	if c.config.NotificationHandler != nil {
		c.config.NotificationHandler(notification)
	}
}

func (c *ManualKafkaConsumer) invokeErrorHandler(err error) {
	if c.config.ErrorHandler != nil {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.ErrorHandlingTimeout)
		defer cancel()

		resultCh := make(chan interface{})
		go func() {
			defer func() {
				if r := recover(); r != nil {
					c.errorf("Panic: While processing %v, trace : %s", r, string(debug.Stack()))
				}
			}()
			c.config.ErrorHandler(ctx, err, nil)
			close(resultCh)
		}()
		select {
		case <-ctx.Done():
			c.debugf("Error Handling TimedOut for err=%v", err)
			return
		case <-resultCh:
			return
		}

	} else {
		c.debugf("Consumer failed to consume with : %+v", string(debug.Stack()))
	}
}
