package jobprocessor

// JobContext contains the JobContext that the caller needs to pass.
type JobContext struct {

	// Timeout is max time allowed to each worker for run.
	Timeout int

	// PoolSize is count of workers the consumer want in workerpool.
	PoolSize int

	// TransactionID is the id for each job transaction to uniquely identify workers of that job.
	TransactionID string
}
