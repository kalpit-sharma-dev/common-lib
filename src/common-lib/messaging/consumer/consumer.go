package consumer

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gammazero/workerpool"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

//go:generate mockgen -package consumer -source=consumer.go -destination=consumer_mock.go .

const (
	// OffsetNewest stands for the log head offset, i.e. the offset that will be
	// assigned to the next message that will be produced to the partition. You
	// can send this to a client's GetOffset method to get this offset, or when
	// calling ConsumePartition to start consuming new messages.
	OffsetNewest int64 = -1
	// OffsetOldest stands for the oldest offset available on the broker for a
	// partition. You can send this to a client's GetOffset method to get this
	// offset, or when calling ConsumePartition to start consuming from the
	// oldest offset that is still available on the broker.
	OffsetOldest int64 = -2
	// OnPull means offset will be committed on Pull before message processing starts.
	OnPull commitMode = -1
	// OnMessageCompletion means offset will be committed on after message processing completes.
	OnMessageCompletion commitMode = -2
	// PullUnOrdered - Messages will be processed in the batches without any order.
	PullUnOrdered consumerMode = -1
	// PullOrdered - Note that message will be processed in per partition sequential Order.
	PullOrdered consumerMode = -2
	/*
		//PullOrderedWithOffsetReplay -Note that message will be processed in per partition sequential Order
		//and the offset can be reset based on the data returned by the function mentioned in the config:HandleCustomOffsetStash
		PullOrderedWithOffsetReplay consumerMode = -3
	*/
)

// ConsumerLogger specifies the interface for all consumer log operations.
type ConsumerLogger interface {
	Printf(format string, v ...interface{})
}

// consumer describes message consumer.
// copy of kafka.Consumer with methods used in this package for mocking/testing.
type consumer interface {
	// SubscribeTopics subscribes to the provided list of topics.
	// This replaces the current subscription.
	SubscribeTopics(topics []string, rebalanceCb kafka.RebalanceCb) error
	// Close Consumer instance.
	// The object is no longer usable after this call.
	Close() error
	// Poll the consumer for messages or events.
	//
	// Will block for at most timeoutMs milliseconds
	//
	// The following callbacks may be triggered:
	//   Subscribe()'s rebalanceCb
	//
	// Returns nil on timeout, else an Event
	Poll(timeoutMs int) (event kafka.Event)
	// StoreOffsets stores the provided list of offsets that will be committed
	// to the offset store according to `auto.commit.interval.ms` or manual
	// offset-less Commit().
	//
	// Returns the stored offsets on success. If at least one offset couldn't be stored,
	// an error and a list of offsets is returned. Each offset can be checked for
	// specific errors via its `.Error` member.
	StoreOffsets(offsets []kafka.TopicPartition) (storedOffsets []kafka.TopicPartition, err error)
	// Assign an atomic set of partitions to consume.
	// This replaces the current assignment.
	Assign(partitions []kafka.TopicPartition) (err error)
	// Unassign the current set of partitions to consume.
	Unassign() (err error)
	// Assignment returns the current partition assignments
	Assignment() (partitions []kafka.TopicPartition, err error)
	// Pause consumption for the provided list of partitions
	//
	// Note that messages already enqueued on the consumer's Event channel
	// (if `go.events.channel.enable` has been set) will NOT be purged by
	// this call, set `go.events.channel.size` accordingly.
	Pause(partitions []kafka.TopicPartition) (err error)
	// Resume consumption for the provided list of partitions
	Resume(partitions []kafka.TopicPartition) (err error)
	// CommitOffsets commits the provided list of offsets
	// This is a blocking call.
	// Returns the committed offsets on success.
	CommitOffsets(offsets []kafka.TopicPartition) ([]kafka.TopicPartition, error)
}

// Service interface.
type Service interface {
	// Pull message from Kafka topic.
	Pull()
	// MarkOffset Update offset for partition and topic.
	MarkOffset(topic string, partition int32, offset int64)
	// Close consumer connection.
	Close() error
	// CloseWait close consumer connection and wait for queued messages to be processed.
	CloseWait() error
	// Health
	Health() (Health, error)
}

// PausableService interface is an extension of Service.
type PausableService interface {
	Service
	PauseResumer
}

type PauseResumer interface {
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

// KafkaConsumer processes available kafka messages.
type KafkaConsumer struct {
	config           *Config
	closing          chan bool
	consumer         consumer
	consumerStrategy consumerStrategy
	commitStrategy   commitStrategy

	healthMx *sync.RWMutex
	health   *Health

	errorLog ConsumerLogger
	infoLog  ConsumerLogger
	debugLog ConsumerLogger

	partMx           *sync.Mutex // it's rarely used and always used to mutate the state so there is no point in using RWMutex
	pausedPartitions map[partition]kafka.TopicPartition
}

type partition struct {
	topic     string
	partition int32
}

type consumerStrategy interface {
	handleAssignedPartitions(p kafka.AssignedPartitions, msgProc func(*Message))
	handleRevokedPartitions(p kafka.RevokedPartitions)
	handleMessage(m *kafka.Message) error
	close(wait bool)
}

type pullOrdered struct {
	partMx             *sync.RWMutex
	partitionConsumers map[partition]partitionConsumer
}

// should not be reused after calling close method.
type partitionConsumer struct {
	requests         chan *Message
	processingDoneWg *sync.WaitGroup
}

// instance should not be reused after calling this method.
func (c partitionConsumer) close(wait bool) {
	close(c.requests)
	if !wait {
		for range c.requests {
			// drain
		}
		return
	}
	c.processingDoneWg.Wait()
}

func newPullOrdered() *pullOrdered {
	return &pullOrdered{
		partMx:             &sync.RWMutex{},
		partitionConsumers: make(map[partition]partitionConsumer),
	}
}

func (po *pullOrdered) handleAssignedPartitions(p kafka.AssignedPartitions, msgProc func(*Message)) {
	po.partMx.Lock()
	defer po.partMx.Unlock()
	for p, messages := range po.partitionConsumers {
		messages.close(false)
		delete(po.partitionConsumers, p)
	}
	for _, part := range p.Partitions {
		p := partition{
			topic:     *part.Topic,
			partition: part.Partition,
		}
		ch := make(chan *Message)
		processingDoneWg := &sync.WaitGroup{}
		po.partitionConsumers[p] = partitionConsumer{
			requests:         ch,
			processingDoneWg: processingDoneWg,
		}
		processingDoneWg.Add(1)
		go func(mCh <-chan *Message, processingDoneWg *sync.WaitGroup, msgProc func(*Message)) {
			defer processingDoneWg.Done()
			for message := range mCh {
				msgProc(message)
			}
		}(ch, processingDoneWg, msgProc)
	}
}

func (po *pullOrdered) handleRevokedPartitions(_ kafka.RevokedPartitions) {
	po.partMx.Lock()
	defer po.partMx.Unlock()
	for p, messages := range po.partitionConsumers {
		messages.close(false)
		delete(po.partitionConsumers, p)
	}
}

func (po *pullOrdered) handleMessage(m *kafka.Message) error {
	p := partition{
		topic:     *m.TopicPartition.Topic,
		partition: m.TopicPartition.Partition,
	}
	po.partMx.RLock()
	defer po.partMx.RUnlock()
	cons, ok := po.partitionConsumers[p]

	if !ok {
		return fmt.Errorf("cannot put message %q from partition %+v into the processing queue as it was closed", m.Key, p)
	}
	cons.requests <- newMessage(m)
	return nil
}

func (po *pullOrdered) close(wait bool) {
	po.partMx.Lock()
	defer po.partMx.Unlock()
	for p, messages := range po.partitionConsumers {
		messages.close(wait)
		delete(po.partitionConsumers, p)
	}
}

type pullUnOrdered struct {
	queueSize int

	workerPool *workerpool.WorkerPool

	queueMx *sync.RWMutex
	queue   chan *Message // One channel for all partitions. Recreated on each rebalance.
}

func newPullUnOrdered(bufferSize int, subscriberPerCore int) *pullUnOrdered {
	poolSize := subscriberPerCore * runtime.NumCPU()
	workerPool := workerpool.New(poolSize)

	return &pullUnOrdered{
		queueSize:  bufferSize,
		workerPool: workerPool,

		queueMx: &sync.RWMutex{},
	}
}

func (pu *pullUnOrdered) handleAssignedPartitions(_ kafka.AssignedPartitions, msgFunc func(*Message)) {
	pu.queueMx.Lock()
	defer pu.queueMx.Unlock()
	if pu.queue != nil {
		close(pu.queue)
	}
	pu.queue = make(chan *Message, pu.queueSize)

	go func(q <-chan *Message, wp *workerpool.WorkerPool, msgFunc func(*Message)) {
		for message := range q {
			m := message
			if wp.Stopped() {
				return
			}
			wp.Submit(func() {
				msgFunc(m)
			})
		}
	}(pu.queue, pu.workerPool, msgFunc)
}

func (pu *pullUnOrdered) handleRevokedPartitions(_ kafka.RevokedPartitions) {
	pu.queueMx.Lock()
	defer pu.queueMx.Unlock()
	if pu.queue != nil {
		close(pu.queue)
	}
	pu.queue = nil
}

func (pu *pullUnOrdered) handleMessage(m *kafka.Message) error {
	pu.queueMx.RLock()
	defer pu.queueMx.RUnlock()
	if pu.queue == nil {
		return fmt.Errorf("cannot put message %q from Topic %q and partition %q into the processing queue as it was closed",
			m.Key, *m.TopicPartition.Topic, m.TopicPartition.Partition)
	}
	pu.queue <- newMessage(m)
	return nil
}

func (pu *pullUnOrdered) close(wait bool) {
	pu.queueMx.Lock()
	defer pu.queueMx.Unlock()
	if pu.queue != nil {
		close(pu.queue)
		if !wait {
			for range pu.queue {
				// purge messages
			}
		}
	}
	pu.queue = nil

	if wait {
		pu.workerPool.StopWait()
	} else {
		pu.workerPool.Stop()
	}
}

// Pull starts polling kafka messages (blocks)
func (c *KafkaConsumer) Pull() {
	poller := make(chan kafka.Event)
	go func() {
		for {
			select {
			case wait := <-c.closing:
				c.infof("Closing receiving events; wait: %v", wait)
				close(poller)
				c.consumerStrategy.close(wait)
				if err := c.consumer.Close(); err != nil {
					c.invokeErrorHandler(err, nil)
				}
				return
			default:
				m := c.consumer.Poll(150) // the higher the value the lower CPU usage when there are no events to consume, but it negatively affects the shutdown time
				if m != nil {
					poller <- m // can be blocked if we receive a message from partition from which we're already processing another message. Consider using buffered channels in consumerStrategies or multiple consumers per topic
				}
			}
		}
	}()

	for ev := range poller {
		switch e := ev.(type) {
		case *kafka.Message:
			if err := c.consumerStrategy.handleMessage(e); err != nil {
				c.invokeErrorHandler(err, newMessage(e))
			}
		case kafka.Error:
			// Errors should generally be considered as informational, the client will try to automatically recover
			if e.Code() == kafka.ErrAllBrokersDown {
				func() {
					c.healthMx.Lock()
					defer c.healthMx.Unlock()
					c.health.ConnectionState = false
				}()
			}
			c.invokeErrorHandler(e, nil)
		// uncomment and send to an OffsetsCommittedHandler if we want to give this info to client
		// case kafka.OffsetsCommitted:
		// invokeNotificationHandler(fmt.Sprintf("%+v", e), c.config)
		// catchall
		default:
			// fmt.Fprintf(os.Stderr, "%% Unhandled event %T ignored: %v\n", e, e)
		}
	}
}

func (c *KafkaConsumer) process(message *Message) {
	c.commitStrategy.onPull(message.GetTransactionID(), message.Topic, message.Partition, message.Offset)
	c.processMessage(message, c.commitStrategy, c.config)
}

// MarkOffset marks offset for partition
func (c *KafkaConsumer) MarkOffset(topic string, partition int32, offset int64) {
	// consumer starts at N, so we have to store N + 1
	// https://docs.confluent.io/5.5.0/clients/librdkafka/rdkafka_8h.html#a047b1e21236fba30898c7c563c2c6777
	_, err := c.consumer.StoreOffsets([]kafka.TopicPartition{{
		Topic:     &topic,
		Partition: int32(partition),
		Offset:    kafka.Offset(offset + 1),
	}})
	if err != nil {
		c.debugf("failed to store topic: %s at %d/%d ==> %+v\n", topic, partition, offset, err)
	}
}

func (c *KafkaConsumer) handleAssignedPartitions(p kafka.AssignedPartitions) {
	if err := c.ResumeAll(); err != nil {
		c.invokeErrorHandler(err, nil)
	}

	if err := c.consumer.Assign(p.Partitions); err != nil {
		c.invokeErrorHandler(err, nil)
		return
	}
	func() {
		c.healthMx.Lock()
		defer c.healthMx.Unlock()
		c.health.ConnectionState = true
	}()
	c.consumerStrategy.handleAssignedPartitions(p, c.process)
}

func (c *KafkaConsumer) handleRevokedPartitions(p kafka.RevokedPartitions) {
	if err := c.ResumeAll(); err != nil {
		c.invokeErrorHandler(err, nil)
	}
	if err := c.consumer.Unassign(); err != nil {
		c.invokeErrorHandler(err, nil)
	}
	c.consumerStrategy.handleRevokedPartitions(p)
}

// Close stops pulling messages from kafka
func (c *KafkaConsumer) Close() error {
	c.partMx.Lock()
	defer c.partMx.Unlock()
	for part := range c.pausedPartitions {
		delete(c.pausedPartitions, part)
	}
	return c.close(false)
}

// CloseWait stops pulling messages from kafka and waits for queued messages to be processed
func (c *KafkaConsumer) CloseWait() error {
	c.partMx.Lock()
	defer c.partMx.Unlock()
	for part := range c.pausedPartitions {
		delete(c.pausedPartitions, part)
	}
	return c.close(true)
}

// Health returns consumer's health
func (c *KafkaConsumer) Health() (Health, error) {
	c.healthMx.RLock()
	defer c.healthMx.RUnlock()
	return *c.health, nil
}

// Pause consumption from the specified topic.
// Use with caution in PullUnOrdered mode as your worker pool can be filled with messages from this partition and prevent consumption of others.
func (c *KafkaConsumer) Pause(topic string, part int32, offset int64) error {
	tp := kafka.TopicPartition{
		Topic:     &topic,
		Partition: part,
		Offset:    kafka.Offset(offset),
	}
	p := partition{
		topic:     *tp.Topic,
		partition: tp.Partition,
	}
	c.partMx.Lock()
	defer c.partMx.Unlock()
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
func (c *KafkaConsumer) Resume(topic string, part int32, offset int64) error {
	tp := kafka.TopicPartition{
		Topic:     &topic,
		Partition: part,
		Offset:    kafka.Offset(offset),
	}
	p := partition{
		topic:     *tp.Topic,
		partition: tp.Partition,
	}
	c.partMx.Lock()
	defer c.partMx.Unlock()
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
func (c *KafkaConsumer) PauseAll() error {
	c.partMx.Lock()
	defer c.partMx.Unlock()
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
func (c *KafkaConsumer) ResumeAll() error {
	c.partMx.Lock()
	defer c.partMx.Unlock()
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

func (c *KafkaConsumer) close(wait bool) error {
	c.closing <- wait
	close(c.closing)
	return nil
}

func (c *KafkaConsumer) errorf(format string, v ...interface{}) {
	if c.errorLog != nil {
		c.errorLog.Printf(format, v...)
	}
}
func (c *KafkaConsumer) infof(format string, v ...interface{}) {
	if c.infoLog != nil {
		c.infoLog.Printf(format, v...)
	}
}
func (c *KafkaConsumer) debugf(format string, v ...interface{}) {
	if c.debugLog != nil {
		c.debugLog.Printf(format, v...)
	}
}

// New creates a new Consumer instance with config settings
func New(config *Config) (*KafkaConsumer, error) {
	if len(config.Address) == 0 || config.Group == "" || len(config.Topics) == 0 || (config.MessageHandler == nil && config.PausableMessageHandler == nil) {
		return nil, fmt.Errorf("some of the values are blank. MessageHandler, Address : %v, Group : %s and Topics : %v",
			config.Address, config.Group, config.Topics)
	}
	if config.MessageHandler != nil && config.PausableMessageHandler != nil {
		return nil, fmt.Errorf("both MessageHandler and PausableMessageHandler are provided but only one is expected")
	}
	if config.ConsumerMode == PullUnOrdered && config.CommitMode == OnMessageCompletion {
		return nil, errors.New("ConsumerMode 'PullUnOrdered' cannot be used with CommitMode 'OnMessageCompletion'")
	}
	configMap := kafka.ConfigMap{
		"bootstrap.servers":               strings.Join(config.Address, ","),
		"group.id":                        config.Group,
		"go.application.rebalance.enable": true,
		"auto.offset.reset":               getResetPosition(config),
		"enable.auto.commit":              true,
		"auto.commit.interval.ms":         config.CommitIntervalMs,
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

	var cs consumerStrategy
	switch config.ConsumerMode {
	case PullOrdered:
		cs = newPullOrdered()
	case PullUnOrdered:
		cs = newPullUnOrdered(config.MaxQueueSize, config.SubscriberPerCore)
	default:
		return nil, fmt.Errorf("unsupported consumer mode %q", config.ConsumerMode)
	}

	kc := &KafkaConsumer{
		config:           config,
		closing:          make(chan bool, 1),
		consumer:         c,
		consumerStrategy: cs,

		healthMx: &sync.RWMutex{},
		health:   newHealth(),

		errorLog: config.ErrorLog,
		infoLog:  config.InfoLog,
		debugLog: config.DebugLog,

		partMx:           &sync.Mutex{},
		pausedPartitions: make(map[partition]kafka.TopicPartition),
	}
	func() {
		kc.healthMx.Lock()
		defer kc.healthMx.Unlock()
		kc.health.Group = config.Group
		kc.health.Topics = config.Topics
		kc.health.Address = config.Address
	}()
	kc.commitStrategy = getCommitStrategy(config, kc.MarkOffset)

	rebalanceCb := func(cons *kafka.Consumer, ev kafka.Event) error {
		switch e := ev.(type) {
		case kafka.AssignedPartitions:
			kc.invokeNotificationHandler(fmt.Sprintf("%+v", e))
			kc.handleAssignedPartitions(e)
			return nil
		case kafka.RevokedPartitions:
			kc.invokeNotificationHandler(fmt.Sprintf("%+v", e))
			kc.handleRevokedPartitions(e)
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

func (c *KafkaConsumer) logKafkaLogs(ch chan kafka.LogEvent) {
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

func getResetPosition(config *Config) string {
	switch config.OffsetsInitial {
	case OffsetNewest:
		return "latest"
	case OffsetOldest:
		return "earliest"
	}
	// TODO: handle custom initial offset?
	return "latest"
}

func (c *KafkaConsumer) processMessage(message *Message, commitStrategy commitStrategy, config *Config) {
	transactionID := message.GetTransactionID()
	commitStrategy.beforeHandler(transactionID, message.Topic, message.Partition, message.Offset)

	var retryCount int64
	var err error

	for retryCount = 0; retryCount <= config.RetryCount; retryCount++ {
		resultErr := func() error {
			ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
			defer cancel()

			select {
			case <-ctx.Done():
				c.debugf("Topic: %s at %d/%d ==> %v\n", message.Topic, message.Partition, message.Offset, ctx.Err())
				return nil
			default:
				errc := make(chan error, 1)
				go func() {
					errc <- c.invokeMessageHandler(ctx, message)
					close(errc)
				}()
				select {
				case <-ctx.Done():
					c.debugf("Topic: %s at %d/%d ==> %v\n", message.Topic, message.Partition, message.Offset, ctx.Err())
					return nil
				case err = <-errc:
					return err
				}
			}
		}()
		if resultErr != nil {
			err = resultErr
			time.Sleep(config.RetryDelay)
		} else {
			err = nil
			break
		}
	}
	commitStrategy.afterHandler(transactionID, message.Topic, message.Partition, message.Offset)

	// queue.processed(message)
	if err != nil {
		c.invokeErrorHandler(err, message)
	}

}

func (c *KafkaConsumer) invokeMessageHandler(ctx context.Context, message *Message) error {
	defer func() {
		if r := recover(); r != nil {
			c.invokeErrorHandler(
				fmt.Errorf("invokeMessageHandler.Panic: While processing %v, trace : %s", r, string(debug.Stack())),
				message,
			)
		}
	}()
	if c.config.MessageHandler != nil {
		return c.config.MessageHandler(ctx, *message)
	}
	return c.config.PausableMessageHandler(ctx, *message, c)
}

func (c *KafkaConsumer) invokeNotificationHandler(notification string) {
	defer func() {
		if r := recover(); r != nil {
			c.invokeErrorHandler(
				fmt.Errorf("invokeNotificationHandler.panic: While processing %v, trace : %s", r, string(debug.Stack())),
				nil,
			)
		}
	}()
	if c.config.NotificationHandler != nil {
		c.config.NotificationHandler(notification)
	}
}

func (c *KafkaConsumer) invokeErrorHandler(err error, message *Message) {
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
			c.config.ErrorHandler(ctx, err, message)
			close(resultCh)
		}()
		select {
		case <-ctx.Done():
			if message != nil {
				c.debugf("Error Handling TimedOut for Topic: %s at %d/%d ==> %v", message.Topic, message.Partition, message.Offset, err)
			} else {
				c.debugf("Error Handling TimedOut for err=%v", err)
			}
			return
		case <-resultCh:
			return
		}

	} else {
		c.debugf("Consumer failed to consume with : %+v", string(debug.Stack()))
	}
}
