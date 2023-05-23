package cassandra

import (
	"fmt"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra/cql"
	exc "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/exception"

	"github.com/gocql/gocql"
)

// Row is a callback function to process database Row
type Row func(map[string]interface{})

// connection is responsible of connecting with Cassandra db
type connection struct {
	session cql.Session
	conf    *DbConfig
}

// Cassandra Protocol version is set to be 4 if you are using cassandra >= 3.0
const protoVersion = 4

// NewDbConnection - returns the struct implementation of DbConnector
func NewDbConnection(conf *DbConfig) (DbConnector, error) {
	return newConnection(conf)
}

// newConnection is a constructor of connection which will initialize struct and will return an open connection object(if no error) of connection
func newConnection(conf *DbConfig) (*connection, error) {
	err := validate(conf)
	if err != nil {
		return nil, err
	}

	db := &connection{conf: conf}
	cluster := gocql.NewCluster()
	cluster.Hosts = conf.Hosts
	cluster.ProtoVersion = protoVersion
	cluster.Authenticator = conf.Authenticator
	cluster.SslOpts = conf.SslOpts
	cluster.DisableInitialHostLookup = conf.DisableInitialHostLookup
	cluster.Keyspace = conf.Keyspace
	cluster.Consistency = conf.Consistency
	cluster.Compressor = gocql.SnappyCompressor{}
	cluster.NumConns = conf.NumConn
	cluster.Timeout = conf.TimeoutMillisecond
	cluster.ConnectTimeout = conf.ConnectTimeout

	obs := &cql.Observer{}
	cluster.QueryObserver = obs
	cluster.BatchObserver = obs

	//If there are some error in connecting to the cluster, below method also does a log.Printf() and logs the error

	err = circuit.Do(conf.CommandName, conf.CircuitBreaker.Enabled, func() error {
		db.session, err = cql.CreateSession(cluster)
		return err
	}, nil)

	if err != nil {
		err = exc.New(ErrDbUnableToConnect, err)
	}
	return db, err
}

func (d connection) Insert(query string, value ...interface{}) error {
	return d.executeDmlQuery(query, value...)
}

func (d connection) Update(query string, value ...interface{}) error {
	return d.executeDmlQuery(query, value...)
}

func (d connection) Delete(query string, value ...interface{}) error {
	return d.executeDmlQuery(query, value...)
}

func (d connection) InsertWithScanCas(dest *map[string]interface{}, query string, value ...interface{}) (bool, error) {
	return d.executeScanCasQuery(dest, query, value...)
}

func (d connection) UpdateWithScanCas(dest *map[string]interface{}, query string, value ...interface{}) (bool, error) {
	return d.executeScanCasQuery(dest, query, value...)
}

func (d connection) DeleteWithScanCas(dest *map[string]interface{}, query string, value ...interface{}) (bool, error) {
	return d.executeScanCasQuery(dest, query, value...)
}

// executeScanCasQuery executes a transaction (query with an IF statement).
// If the transaction is successfully executed, it returns true. If the transaction fails
// beacuse the IF condition was not satisfied, it returns false and populates dest(only for
// insert AND update query) with the existing values in cassandra.
// https://godoc.org/github.com/gocql/gocql#Query.MapScanCAS
func (d connection) executeScanCasQuery(dest *map[string]interface{}, query string, value ...interface{}) (bool, error) {
	resultMap := make(map[string]interface{})
	if d.session == nil {
		return false, exc.New(ErrDbNoOpenConnection, nil)
	}
	q := d.session.Query(query, value...)

	isApplied := false
	err := circuit.Do(d.conf.CommandName, d.conf.CircuitBreaker.Enabled, func() error {
		isapplied, err := q.MapScanCAS(resultMap)
		isApplied = isapplied
		return err
	}, nil)

	defer q.Release()
	if err != nil {
		return false, fmt.Errorf("%s : %s ", ErrDbDMLFailed, err.Error())
	}
	//This has been purposefully done here so that after resultMap is populated,
	//its data can be copied to dest
	if dest != nil {
		*dest = resultMap
	}
	return isApplied, nil
}

// executeDmlQueryWithRetry : executes a query with retry policy
// https://github.com/gocql/gocql/blob/master/policies.go
func (d connection) ExecuteDmlWithRetrial(policy gocql.RetryPolicy, query string, value ...interface{}) error {
	if policy == nil {
		return exc.New(ErrPolicyNotDefined, nil)
	}
	return d.executeQuery(policy, query, value...)
}

func (d connection) executeDmlQuery(query string, value ...interface{}) error {
	return d.executeQuery(nil, query, value...)
}

func (d connection) executeQuery(policy gocql.RetryPolicy, query string, value ...interface{}) error {
	if d.session == nil {
		return exc.New(ErrDbNoOpenConnection, nil)
	}
	q := d.session.Query(query, value...)
	if policy != nil {
		q.RetryPolicy(policy)
	}

	// Warp this to Circuit breaker code
	err := circuit.Do(d.conf.CommandName, d.conf.CircuitBreaker.Enabled, q.Exec, nil)
	defer q.Release()
	if err != nil {
		return exc.New(ErrDbDMLFailed, err)
	}
	return nil
}

func (d connection) Select(query string, value ...interface{}) ([]map[string]interface{}, error) {
	q := d.session.Query(query, value...)
	return d.executeSelectQuery(nil, q)
}

func (d connection) Query(query string, value ...interface{}) cql.Query {
	return d.session.Query(query, value...)
}

func (d connection) RunSelectQuery(q cql.Query) ([]map[string]interface{}, error) {
	return d.executeSelectQuery(nil, q)
}

func (d connection) SelectWithRetrial(policy gocql.RetryPolicy, query string, value ...interface{}) ([]map[string]interface{}, error) {
	if policy == nil {
		return nil, exc.New(ErrPolicyNotDefined, nil)
	}
	q := d.session.Query(query, value...)
	return d.executeSelectQuery(policy, q)
}

func (d connection) executeSelectQuery(policy gocql.RetryPolicy, q cql.Query) ([]map[string]interface{}, error) {
	if policy != nil {
		q.RetryPolicy(policy)
	}

	var data []map[string]interface{}

	// Warp all of these to Circuit breaker code
	err := circuit.Do(d.conf.CommandName, d.conf.CircuitBreaker.Enabled, func() error {
		iter := q.Iter()
		var err error
		data, err = iter.SliceMap()
		defer q.Release()
		if err != nil {
			return exc.New(ErrDbUnableToFetchRecord, err)
		}

		err = iter.Close()
		if err != nil {
			return exc.New(ErrDbUnableToFetchRecord, err)
		}

		return nil
	}, nil)

	return data, err
}

func (d connection) SelectWithPaging(page int, callback ProcessRow, query string, value ...interface{}) error {
	q := d.session.Query(query, value...) //.Consistency(gocql.One)
	q.PageSize(page)
	defer q.Release()
	iter := q.Iter()
	m := make(map[string]interface{})
	for {
		if !iter.MapScan(m) {
			break
		}
		callback(m)
		m = make(map[string]interface{})
	}
	return iter.Close()
}

// GetRandomUUID() returns the random generated UUID
func (connection) GetRandomUUID() (string, error) {
	uuid, err := gocql.RandomUUID()
	if err != nil {
		return "", exc.New(ErrUUID, err)
	}
	return uuid.String(), nil
}

// Close function closes the connection and does not return error
func (d connection) Close() {
	if d.session != nil {
		d.session.Close()
	}
}

// Closed function to check is session is closed or not
func (d connection) Closed() bool {
	if d.session != nil {
		return d.session.Closed()
	}
	return true
}
