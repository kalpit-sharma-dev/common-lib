package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

// BatchQueryExecutor is based on classic builder pattern which leverage user to add,
// as many query with ease with different level of code execution unit and
// at the end will Execute instruction will instruct batch to execute al queries
type BatchQueryExecutor interface {
	//AddQuery is add query with its args to the batch and return the same instanceso that user
	//can keep on adding or appending queries with its args and once they are done can
	//instruct execute command to execute all queries at once
	AddQuery(query string, args ...interface{}) BatchQueryExecutor

	//Execute to instruct batch to execute all queries at once
	Execute() error
}

// GetBatchQueryExecutor is factory method to get BatchQueryExecutor
func GetBatchQueryExecutor(conf *DbConfig) (BatchQueryExecutor, error) {
	if conf == nil {
		return nil, fmt.Errorf("config : %+v is mandatory for getting executor of batch", conf)
	}

	connector, err := getBatchConnection(conf)
	if err != nil {
		return nil, err
	}
	batchConnection := connector.(*batchConnection)
	batch := batchConnection.Connection.session.NewBatch(gocql.LoggedBatch)
	return &batchQueryExecutorImpl{batchConnection, batch}, nil
}

type batchQueryExecutorImpl struct {
	batchConnection *batchConnection
	batch           *gocql.Batch
}

var bSession batchCassandraSession
var bSessionInitialized = false
var batchConnInstance BatchDbConnector

func getBatchConnection(conf *DbConfig) (BatchDbConnector, error) {
	if !bSessionInitialized || bSession.Closed() {
		mu.Lock()
		defer mu.Unlock()
		if !bSessionInitialized || bSession.Closed() {
			if bSessionInitialized {
				bSession.closeSuper()
			}
			batchConn, err := NewBatchDbConnection(conf)
			if err != nil {
				return nil, err
			}
			batchConnInstance = batchConn
			bSession = batchCassandraSession{batchConn}
			bSessionInitialized = true
		}
	}
	return batchConnInstance, nil
}

func (b *batchQueryExecutorImpl) AddQuery(query string, args ...interface{}) BatchQueryExecutor {
	b.batch.Query(query, args...)
	return b
}

func (b *batchQueryExecutorImpl) Execute() error {
	err := circuit.Do(b.batchConnection.Connection.conf.CommandName, b.batchConnection.Connection.conf.CircuitBreaker.Enabled, func() error {
		err := b.batchConnection.Connection.session.ExecuteBatch(b.batch)
		return err
	}, nil)
	return err
}
