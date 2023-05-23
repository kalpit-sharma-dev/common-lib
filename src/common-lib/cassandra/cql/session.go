package cql

import (
	"context"
	"errors"

	"github.com/gocql/gocql"
)

//go:generate mockgen -package mock -destination=mock/mocks.go . Session,Query

// Session - Interface on top of Cassandra session object
// So that we can mock the behavior and unit test wrapper
type Session interface {
	ExecuteBatch(*gocql.Batch) error
	NewBatch(gocql.BatchType) *gocql.Batch
	Query(stmt string, values ...interface{}) Query
	Close()
	Closed() bool
}

// Query - Interface on top of Cassandra query object
// So that we can mock the behavior and unit test wrapper
type Query interface {
	Statement() string
	String() string
	Attempts() int
	Latency() int64
	Consistency(c gocql.Consistency) *gocql.Query
	GetConsistency() gocql.Consistency
	SetConsistency(c gocql.Consistency)
	Trace(trace gocql.Tracer) *gocql.Query
	Observer(observer gocql.QueryObserver) *gocql.Query
	PageSize(n int) *gocql.Query
	DefaultTimestamp(enable bool) *gocql.Query
	WithTimestamp(timestamp int64) *gocql.Query
	RoutingKey(routingKey []byte) *gocql.Query
	WithContext(ctx context.Context) *gocql.Query
	Keyspace() string
	GetRoutingKey() ([]byte, error)
	Prefetch(p float64) *gocql.Query
	RetryPolicy(r gocql.RetryPolicy) *gocql.Query
	IsIdempotent() bool
	Idempotent(value bool) *gocql.Query
	Bind(v ...interface{}) *gocql.Query
	SerialConsistency(cons gocql.SerialConsistency) *gocql.Query
	PageState(state []byte) *gocql.Query
	NoSkipMetadata() *gocql.Query
	Exec() error
	Iter() *gocql.Iter
	MapScan(m map[string]interface{}) error
	Scan(dest ...interface{}) error
	ScanCAS(dest ...interface{}) (applied bool, err error)
	MapScanCAS(dest map[string]interface{}) (applied bool, err error)
	Release()
}

type sessionimpl struct {
	*gocql.Session
}

// CreateSession - A constructor of connection which will initialize struct
// and return an open connection object (if no error) of connection
var CreateSession = func(cluster *gocql.ClusterConfig) (Session, error) {
	instance := &sessionimpl{}

	if cluster == nil {
		return instance, errors.New("ErrorNilCluster")
	}

	obs := &Observer{}
	cluster.QueryObserver = obs
	cluster.BatchObserver = obs

	var err error
	//If there are some error in connecting to the cluster, below method also does a log.Printf() and logs the error
	instance.Session, err = cluster.CreateSession()
	return instance, err
}

func (s *sessionimpl) Query(stmt string, values ...interface{}) Query {
	return s.Session.Query(stmt, values...)
}
