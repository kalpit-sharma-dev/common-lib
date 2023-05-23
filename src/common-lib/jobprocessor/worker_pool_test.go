package jobprocessor

import (
	"context"
	"sync"
	"testing"
)

type WorkerNotPanic struct {
	val  string
	val2 string
}

type WorkerPanic struct {
	val1  string
	val21 string
}

func (i *WorkerNotPanic) Work(id int) error {
	return nil
}

func (i *WorkerPanic) Work(id int) error {
	panic("it is panicing")
}

func TestWorkerPool_Shutdown(t *testing.T) {
	pool, _ := Initialize(JobContext{PoolSize: 100})
	wg := &sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		input := WorkerNotPanic{
			val:  "Workerpool",
			val2: "Job",
		}
		go func() {
			pool.AddJob(&input)
			wg.Done()
		}()
	}
	wg.Wait()

	tests := []struct {
		name string
		w    *WorkerPool
	}{
		{
			name: "Shutdown worker pool workpool with 100 routines and 1000 job submitted",
			w:    pool,
		},
	}

	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			tt.w.Shutdown()
			if pool.counter != 0 {
				t.Errorf("TestWorkerPool_Shutdown failed:number of jobs not shutdown %d", pool.counter)
			}

		})
	}
}

func TestWorkerPool_AddJob(t *testing.T) {
	type args struct {
		work Worker
	}

	w, _ := Initialize(JobContext{PoolSize: 5})

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Worker should add job to workerpool and return true",

			args: args{
				work: &WorkerNotPanic{
					val:  "11",
					val2: "11",
				},
			},
		},
		{
			name: "Worker should add job to workerpool and return true",

			args: args{
				work: &WorkerNotPanic{
					val:  "12",
					val2: "12",
				},
			},
		},
		{
			name: "Worker should add job to workerpool and return true",

			args: args{
				work: &WorkerNotPanic{
					val:  "13",
					val2: "13",
				},
			},
		},
		{
			name: "Worker should add job to workerpool and return true",

			args: args{
				work: &WorkerNotPanic{
					val:  "14",
					val2: "14",
				},
			},
		},
		{
			name: "Worker should add job to workerpool and return true",

			args: args{
				work: &WorkerNotPanic{
					val:  "15",
					val2: "15",
				},
			},
		},
		{
			name: "Worker should add job to workerpool and return true",

			args: args{
				work: &WorkerNotPanic{
					val:  "16",
					val2: "16",
				},
			},
		},
		{
			name: "Worker should add job to workerpool and return true",

			args: args{
				work: &WorkerNotPanic{
					val:  "17",
					val2: "17",
				},
			},
		},
		{
			name: "panic recovery scenario:Worker should add job to workerpool and return true",

			args: args{
				work: &WorkerPanic{
					val1:  "18",
					val21: "18",
				},
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			w.AddJob(tt.args.work)
		})
	}
}

func TestWorkerPool_work(t *testing.T) {
	type fields struct {
		jobContext JobContext
	}

	w := &WorkerPool{}

	workerNotPanic := WorkerNotPanic{
		val:  "123",
		val2: "345",
	}

	workerPanic := WorkerPanic{
		val1:  "123",
		val21: "345",
	}

	type args struct {
		work Worker
		id   int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Should execute the provided work without error",
			args: args{
				work: &workerNotPanic,
			},
			fields: fields{
				jobContext: JobContext{
					PoolSize: 10,
					Timeout:  60,
				},
			},
		},
		{
			name: "Panic recovery scenario:Should execute the provided work without error",
			args: args{
				work: &workerPanic,
			},
			fields: fields{
				jobContext: JobContext{
					PoolSize: 1,
					Timeout:  6,
				},
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			w.work(tt.args.work, tt.args.id, tt.fields.jobContext)
		})
	}
}

func TestWorkerPool_Initialize(t *testing.T) {
	type fields struct {
		jobsChan      chan Worker
		workChan      chan Worker
		ctx           context.Context
		cancelFunc    context.CancelFunc
		wg            *sync.WaitGroup
		counter       int64
		jobContext    JobContext
		transactionID string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Creating Workerpool by providing jobContext-shouldnot return error",
			fields: fields{
				jobContext: JobContext{
					PoolSize: 10,
					Timeout:  10,
				},
			},
			wantErr: false,
		},
		{
			name:    "Creating Workerpool with default jobContext-shouldnot return error",
			wantErr: false,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			_, err := Initialize(tt.fields.jobContext)
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkerPool.Initialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
