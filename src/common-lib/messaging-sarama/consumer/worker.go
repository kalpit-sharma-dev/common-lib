package consumer

import (
	"context"
	"runtime"
	"sync"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
)

type workerPool struct {
	//pool *work.Pool
	cfg        Config
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         *sync.WaitGroup
	pool       subscriber
}

func (w *workerPool) initialize() {
	poolSize := runtime.NumCPU() * w.cfg.SubscriberPerCore
	Logger().Info(utils.GetTransactionID(), "Worker pool size : %d", poolSize)

	w.pool = subscriber{
		ingestChan: make(chan job, 1),
		jobsChan:   make(chan job, poolSize),
	}
	// Set up cancellation context and waitgroup
	w.ctx, w.cancelFunc = context.WithCancel(context.Background())
	w.wg = &sync.WaitGroup{}

	// Start consumer with cancellation context passed
	go w.pool.startConsumer(w.ctx)

	// Start workers and Add [workerPoolSize] to WaitGroup
	w.wg.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		go w.pool.workerFunc(w.wg, i)
	}
}

func (w *workerPool) addJob(message Message, commitStrategy commitStrategy) {
	w.pool.callbackFunc(job{
		cfg:            w.cfg,
		message:        message,
		commitStrategy: commitStrategy,
	})
}

func (w *workerPool) Shutdown() {
	Logger().Info("WorkerPool", "*********************************\nShutdown request received\n*********************************")
	w.cancelFunc() // Signal cancellation to context.Context
	w.wg.Wait()    // Block here until are workers are done
	Logger().Info("WorkerPool", "All workers done, shutting down")
	Logger().Info("WorkerPool", "*********************************\nGraceful shutdown completed\n*********************************")
}
