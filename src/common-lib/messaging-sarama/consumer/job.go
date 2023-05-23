package consumer

import (
	"context"
	"fmt"
	"runtime/debug"
)

type job struct {
	cfg            Config
	message        Message
	commitStrategy commitStrategy
}

func (j job) Work(id int) {
	defer func() {
		err := recover()
		if err != nil {
			j.cfg.ErrorHandler(err.(error), &j.message)
			Logger().Error(j.message.transactionID, "JobPanicRecovered", "Recovered job,error %v,Stack Trace : %s", err, debug.Stack())
		}
	}()

	transactionID := j.message.GetTransactionID()
	Logger().Debug(transactionID, "partition id %d, worker pool id %d started", j.message.Partition, id)
	j.commitStrategy.beforeHandler(transactionID, j.message.Topic, j.message.Partition, j.message.Offset)

	ctx, cancel := context.WithTimeout(context.Background(), j.cfg.Timeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- invokeMessageHandler(j.message, j.cfg)
	}()

	select {
	case <-ctx.Done():
		Logger().Debug(transactionID, "Context timeout, Topic: %s at %d/%d ==> %v\n", j.message.Topic, j.message.Partition, j.message.Offset, ctx.Err())
		j.commitStrategy.afterHandler(transactionID, j.message.Topic, j.message.Partition, j.message.Offset)
	case err := <-done:
		j.commitStrategy.afterHandler(transactionID, j.message.Topic, j.message.Partition, j.message.Offset)
		Logger().Debug(transactionID, "Processed, partition id %d, worker pool id %d", j.message.Partition, id)
		if err != nil && j.cfg.ErrorHandler != nil {
			j.cfg.ErrorHandler(err, &j.message)
		}
	}
	Logger().Debug(transactionID, "partition id %d, worker pool id %d done", j.message.Partition, id)
}

func invokeMessageHandler(message Message, cfg Config) error {
	defer func() {
		if r := recover(); r != nil {
			invokeErrorHandler(
				fmt.Errorf("invokeMessageHandler.Panic: While processing %v, trace : %s", r, string(debug.Stack())),
				&message,
				cfg,
			)
		}
	}()
	return cfg.MessageHandler(message)
}
