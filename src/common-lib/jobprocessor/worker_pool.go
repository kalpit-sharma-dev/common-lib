package jobprocessor

import (
	"context"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/exec/with"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

const (
	workRoutineCountDefault   = 10
	workRoutineTimeOutDefault = 60
)

// Logger initialized with  Discard Logger.
var Logger = logger.DiscardLogger

// Worker needs to be implemented by the consumers.
type Worker interface {
	Work(id int) error
}

// WorkerPool contains the fields required to initialize a workerpool.
type WorkerPool struct {
	jobsChan   chan Worker
	workChan   chan Worker
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         *sync.WaitGroup
	counter    int64
}

// Initialize needs to be called by the consumers to initialze a wroker pool.
func Initialize(ctx JobContext) (*WorkerPool, error) {
	workpool := WorkerPool{
		jobsChan: make(chan Worker),
		workChan: make(chan Worker),
	}
	// Set up cancellation context and waitgroup
	workpool.ctx, workpool.cancelFunc = context.WithCancel(context.Background())

	workpool.wg = &sync.WaitGroup{}

	jobContext := verifyJobConfiguration(ctx)

	// Start consumer with cancellation context passed
	poolSize := jobContext.PoolSize

	go workpool.process(workpool.ctx)

	workpool.wg.Add(poolSize)

	for r := 1; r <= poolSize; r++ {
		go workpool.worker(r, workpool.wg, jobContext)
	}

	return &workpool, nil
}

func (w *WorkerPool) process(ctx context.Context) {
	defer close(w.jobsChan)

	for {
		select {
		case job := <-w.workChan:
			w.jobsChan <- job
		case <-ctx.Done():
			return
		}
	}
}

func (w *WorkerPool) worker(id int, wg *sync.WaitGroup, jobContext JobContext) {
	defer wg.Done()

	for wrk := range w.jobsChan {
		w.work(wrk, id, jobContext)
	}
}

func (w *WorkerPool) work(work Worker, id int, jobContext JobContext) {
	defer func() {
		err := recover()
		if err != nil {
			Logger().Error(jobContext.TransactionID, "Panic", "Recovered", "error: %v,StackTrace : %s", err, debug.Stack())
		}
	}()

	defer func() {
		atomic.AddInt64(&w.counter, -1)
	}()

	wrkRoutineTimeOut := jobContext.Timeout
	done := make(chan error, 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(wrkRoutineTimeOut))
	defer cancel()

	go with.Recover("Work", jobContext.TransactionID, func() {
		done <- work.Work(id)
	}, func(transaction string, err error) {
		done <- nil
		Logger().Debug(jobContext.TransactionID, "Go routine panic occurred in transaction %s error %v", transaction, err)
	})

	select {
	case <-ctx.Done():
		Logger().Debug(jobContext.TransactionID, "Context timeout", ctx.Err())
	case err := <-done:
		if err != nil {
			Logger().Debug(jobContext.TransactionID, "Error occurred in worker routine", "error %v,Stack Trace : %s", err, debug.Stack())
		}
	}
}

func verifyJobConfiguration(ctx JobContext) JobContext {
	if ctx.Timeout <= 0 {
		ctx.Timeout = workRoutineTimeOutDefault
	}

	if ctx.PoolSize <= 0 {
		ctx.PoolSize = workRoutineCountDefault
	}

	return ctx
}

// AddJob adds a job to the worker pool.
func (w *WorkerPool) AddJob(work Worker) {
	atomic.AddInt64(&w.counter, 1)
	w.workChan <- work
}

// Shutdown needs to be called to shut down the worker pool.
func (w *WorkerPool) Shutdown() {
	w.cancelFunc() // Signal cancellation to context.Context
	w.wg.Wait()    // Block here until are workers are done
}
