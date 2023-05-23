package consumer

import (
	"context"
	"sync"
)

type subscriber struct {
	ingestChan chan job
	jobsChan   chan job
}

// callbackFunc is invoked each time the external lib passes an event to us.
func (s subscriber) callbackFunc(event job) {
	s.ingestChan <- event
}

// workerFunc starts a single worker function that will range on the jobsChan until that channel closes.
func (s subscriber) workerFunc(wg *sync.WaitGroup, index int) {
	defer wg.Done()

	for job := range s.jobsChan {
		job.Work(index)
	}
	Logger().Info("WorkerPool", "worker pool id %d inturppted", index)
}

// startConsumer acts as the proxy between the ingestChan and jobsChan, with a select to support graceful shutdown.
func (s subscriber) startConsumer(ctx context.Context) {
	defer close(s.jobsChan)
	for {
		select {
		case event := <-s.ingestChan:
			select {
			case <-ctx.Done():
				return
			default:
			}
			s.jobsChan <- event
		case <-ctx.Done():
			return
		}
	}
}
