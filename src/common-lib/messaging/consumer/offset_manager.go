package consumer

import (
	"fmt"
	"sync"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

// Logger for consumer.
// Deprecated: in favour of ConsumerLogger.
var Logger = logger.DiscardLogger

type markOffset func(topic string, partition int32, offset int64)

//go:generate mockgen -package consumer -source=offset_manager.go -destination=offset_manager_mock.go .

type commitStrategy interface {
	onPull(transaction, topic string, partition int32, offset int64)
	beforeHandler(transaction, topic string, partition int32, offset int64)
	afterHandler(transaction, topic string, partition int32, offset int64)
}

func getCommitStrategy(cfg *Config, markOffset markOffset) commitStrategy {
	switch cfg.CommitMode {
	case OnPull:
		return &onPull{
			markOffset: markOffset,
			cfg:        cfg,
		}
	case OnMessageCompletion:
		return &onMessageCompletion{
			state:      make(map[string]*partitionState),
			markOffset: markOffset,
			canProcess: make(chan bool, 1),
			cfg:        cfg,
		}
	}

	panic(fmt.Errorf("commitStrategy not available for Commit Mode: %v", cfg.CommitMode))
}

// OnPull Strategy Handler
type onPull struct {
	markOffset markOffset
	cfg        *Config
}

func (o *onPull) onPull(transaction, topic string, partition int32, offset int64) {
	o.markOffset(topic, partition, offset)
}

func (o *onPull) beforeHandler(transaction, topic string, partition int32, offset int64) {

}

func (o *onPull) afterHandler(transaction, topic string, partition int32, offset int64) {
}

// onMessageCompletion Strategy Handler
type onMessageCompletion struct {
	state      map[string]*partitionState
	markOffset markOffset
	canProcess chan bool
	cfg        *Config
	mutex      sync.Mutex
}

func (o *onMessageCompletion) onPull(transaction, topic string, partition int32, offset int64) {
}

func (o *onMessageCompletion) beforeHandler(transaction, topic string, partition int32, offset int64) {
	key := getPartitionKey(topic, partition)
	o.mutex.Lock()
	value, ok := o.state[key]
	o.mutex.Unlock()
	if !ok {
		value = &partitionState{
			offset:             make([]*offsetStatus, 0),
			lastCommitedOffset: -1,
		}
		o.mutex.Lock()
		o.state[key] = value
		o.mutex.Unlock()
	}
	value.setOffset(append(value.offset, &offsetStatus{offset: offset, status: inProgress}))
}

func (o *onMessageCompletion) afterHandler(transaction, topic string, partition int32, offset int64) {
	key := fmt.Sprintf("%s_%d", topic, partition)
	o.mutex.Lock()
	value, ok := o.state[key]
	o.mutex.Unlock()
	if !ok {
		o.markOffset(topic, partition, offset)
		return
	}
	o.process(transaction, topic, partition, offset, value)
}

func (o *onMessageCompletion) process(transaction, topic string, partition int32, offset int64, value *partitionState) {
	log := Logger() // nolint - false positive
	select {
	case o.canProcess <- true:
		{
			commitOffset, commit := value.getCommitOffset(offset)
			log.Debug(transaction, "CommitOffset: %d/%v, lastCommitedOffset: %d, offsets: %+v\n", commitOffset, commit, value.lastCommitedOffset, value.offset)
			if commit {
				o.markOffset(topic, partition, commitOffset)
			}
			<-o.canProcess
		}
	default:
		log.Trace(transaction, "Default called for offset : %d\n", offset)
		value.updateStatus(offset)
	}
}

func getPartitionKey(topic string, partition int32) string {
	return fmt.Sprintf("%s_%d", topic, partition)
}
