package zookeeper

import (
	"context"
	"sync"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed/lock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed/scheduler"
)

// DistributedJobListener initializes distributed job listeners
func (schedulerImpl) DistributedJobListener(ctx context.Context, wg *sync.WaitGroup, jobs []scheduler.DistributedJob, interval time.Duration) error {
	for _, job := range jobs {
		setDistributedJob(ctx, wg, job, interval)
	}
	return nil
}

func setDistributedJob(ctx context.Context, wg *sync.WaitGroup, job scheduler.DistributedJob, interval time.Duration) {
	wg.Add(1)
	locker := NewLock(job.GetName())

	go func() {
		defer func() {
			wg.Done()
			err := locker.Unlock()
			if err != nil {
				Logger().Error(defaultTransaction, "locker.UnlockFailed", "%v", err)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				Logger().Warn(defaultTransaction, "Warning!!! Distributed Job Listener received ctx.Done(): %v", ctx)
				return

			case <-time.After(interval):
				//get lock
				if err := locker.Lock(); err != nil {
					switch getLockErrAction("lock", job, err) {
					case lock.TryLock:
						continue
					case lock.CreateNewLock:
						Logger().Error(defaultTransaction, "locker.LockFailed", "Distributed Job [%s]. Lock failed. Creating a new lock, err: %v", job.GetName(), err)
						locker = NewLock(job.GetName())
						continue
					case lock.TryUnlock:
						//proceed to unlock
					}
				} else {
					process(ctx, job)
				}

				//unlock
				if err := locker.Unlock(); err != nil {
					if getLockErrAction("unlock", job, err) == lock.CreateNewLock {
						Logger().Error(defaultTransaction, "locker.UnlockFailed", "Distributed Job [%s]. Unlock failed. Creating a new lock, err: %v", job.GetName(), err)
						locker = NewLock(job.GetName())
					}
				}
			}
		}
	}()
}

func getLockErrAction(action string, job scheduler.DistributedJob, err error) lock.ErrorAction {
	le, ok := (err).(lock.Error)
	if ok && le.Code != nil {
		Logger().Error(defaultTransaction, "lock.Error", "Distributed Job [%s]. Couldn't make a %s, err: %v", job.GetName(), action, err)
		return le.Action
	}
	return lock.Ignore
}

func process(ctx context.Context, job scheduler.DistributedJob) {
	items, err := Queue.GetList(job.GetName())
	if err != nil {
		Logger().Error(defaultTransaction, "Queue.ListFailed", "Couldn't get notification for distributed job: %v, err: %v", job.GetName(), err)
		return
	}
	if len(items) > 0 {
		processQueue(ctx, items, job)
	}
}

func processQueue(ctx context.Context, items []string, job scheduler.DistributedJob) {
	Logger().Info(defaultTransaction, "Distributed Job [%s]. Found %d notification(s). Executing callback", job.GetName(), len(items))
	var itemsData [][]byte

	for _, item := range items {
		itemData, err := Queue.GetItemData(job.GetName(), item)
		if err != nil {
			Logger().Error(defaultTransaction, "Queue.GetItemDataFailed", "Distributed Job [%s]. Couldn't get job data, err: %v", job.GetName(), err)
			continue
		}

		if len(itemData) != 0 {
			itemsData = append(itemsData, itemData)
		}

		err = Queue.RemoveItem(job.GetName(), item)
		if err != nil {
			Logger().Error(defaultTransaction, "Queue.RemoveItemFailed", "Distributed Job [%s]. Couldn't remove notification from queue, err: %v", job.GetName(), err)
		}
	}

	job.Callback(ctx, itemsData)
}
