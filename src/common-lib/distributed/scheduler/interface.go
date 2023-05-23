package scheduler

import (
	"context"
	"sync"
	"time"
)

type (
	// Interface - Interface to hold Scheduler data
	Interface interface {
		DistributedScheduler(ctx context.Context, wg *sync.WaitGroup, jobs []ScheduledJob, interval time.Duration) error
		DistributedJobListener(ctx context.Context, wg *sync.WaitGroup, jobs []DistributedJob, interval time.Duration) error
	}

	// ScheduledJob interface
	ScheduledJob interface {
		GetName() string
		GetTask() string
		GetSchedule() string
	}

	// DistributedJob interface
	DistributedJob interface {
		GetName() string
		Callback(i ...interface{})
	}
)
