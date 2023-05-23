package zookeeper

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed/queue"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed/scheduler"
)

func TestJob(t *testing.T) {
	job := Job{
		Name:     "Name",
		Schedule: "Schedule",
		Task:     "Task",
	}

	if job.GetName() != job.Name {
		t.Errorf("expected name: %s, got: %s", job.Name, job.GetName())
	}
	if job.GetSchedule() != job.Schedule {
		t.Errorf("expected schedule: %s, got: %s", job.Schedule, job.GetSchedule())
	}
	if job.GetTask() != job.Task {
		t.Errorf("expected task: %s, got: %s", job.Task, job.GetTask())
	}
}

func TestStartScheduler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockQueue := NewMockInterface(ctrl)

	defer func(queue queue.Interface) { Queue = queue }(Queue)
	Queue = mockQueue

	type args struct {
		scenario          string
		job               Job
		existQueueError   error
		queueExists       bool
		createQueueError  error
		expectedEntrySize int
		errorExpected     bool
	}

	tests := []args{
		{
			scenario: "existQueueError",
			job: Job{
				Name:     "job",
				Schedule: "@every 1m",
				Task:     "job",
			},
			existQueueError:   errors.New("injected"),
			errorExpected:     true,
			expectedEntrySize: 0,
		},
		{
			scenario: "createQueueError",
			job: Job{
				Name:     "job",
				Schedule: "@every 1m",
				Task:     "job",
			},
			createQueueError:  errors.New("injected"),
			errorExpected:     true,
			expectedEntrySize: 0,
		},
		{
			scenario: "QueueNotExists",
			job: Job{
				Name:     "job",
				Schedule: "@every 1m",
				Task:     "job",
			},
			errorExpected:     false,
			expectedEntrySize: 1,
		},
		{
			scenario: "QueueExists",
			job: Job{
				Name:     "job",
				Schedule: "@every 1m",
				Task:     "job",
			},
			queueExists:       true,
			errorExpected:     false,
			expectedEntrySize: 1,
		},
		{
			scenario: "cronError",
			job: Job{
				Name:     "job",
				Schedule: "invalid-schedule",
				Task:     "job",
			},
			errorExpected:     true,
			expectedEntrySize: 0,
		},
		{
			scenario: "success",
			job: Job{
				Name:     "job",
				Schedule: "@every 1m",
				Task:     "job",
			},
			expectedEntrySize: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			mockQueue.EXPECT().Exists("job").AnyTimes().DoAndReturn(func(queueName string) (bool, error) {
				return test.queueExists, test.existQueueError
			})
			mockQueue.EXPECT().Create("job").AnyTimes().DoAndReturn(func(queueName string) (string, error) {
				return "", test.createQueueError
			})
			scheduledJobs = []scheduler.ScheduledJob{test.job}

			startScheduler()

			if len(schedulerCron.Entries()) != test.expectedEntrySize {
				t.Fatalf("expected the size of cron jobs %d, found %d", test.expectedEntrySize, len(schedulerCron.Entries()))
			}

			scheduledJobs = nil
		})
	}
}
