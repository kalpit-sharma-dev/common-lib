package cassandra

import (
	"github.com/gocql/gocql"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra/cql"
)

// ProcessRow is a callback function to process database Row
type ProcessRow func(map[string]interface{})

// DbConnector interface is responsible for dealing with the database
// This interface will not expose Open() method to open connection, this will be the responsibility of the underlying implementation.
type DbConnector interface {
	// Insert will insert the record
	Insert(query string, value ...interface{}) error

	// Update will update the record
	Update(query string, value ...interface{}) error

	// Delete will delete the record
	Delete(query string, value ...interface{}) error

	RunSelectQuery(q cql.Query) ([]map[string]interface{}, error)
	Query(query string, value ...interface{}) cql.Query

	// ExecuteDmlWithRetrial executes query with defined retry policy
	ExecuteDmlWithRetrial(policy gocql.RetryPolicy, query string, value ...interface{}) error

	// SelectWithRetrial executes select query with retrial policy
	SelectWithRetrial(policy gocql.RetryPolicy, query string, value ...interface{}) ([]map[string]interface{}, error)

	// InsertWithScanCas executes a lightweight transaction (i.e. an UPDATE or INSERT statement containing an IF clause).
	// If the transaction fails because the existing values did not match, the previous values will be stored in dest.
	InsertWithScanCas(dest *map[string]interface{}, query string, value ...interface{}) (bool, error)

	// UpdateWithScanCas executes a lightweight transaction (i.e. an UPDATE or INSERT statement containing an IF clause).
	// If the transaction fails because the existing values did not match, the previous values will be stored in dest.
	UpdateWithScanCas(dest *map[string]interface{}, query string, value ...interface{}) (bool, error)

	// DeleteWithScanCas executes a lightweight transaction (i.e. an UPDATE or INSERT statement containing an IF clause).
	// If the transaction fails because the existing values did not match, the previous values will be stored in dest.
	DeleteWithScanCas(dest *map[string]interface{}, query string, value ...interface{}) (bool, error)

	// Close method should be called by the caller at the end of all operation to Close the Db connection
	Close()

	// Select gets the record from the table and returns the result set
	Select(query string, value ...interface{}) ([]map[string]interface{}, error)

	// SelectWithPaging gets the record from the table and calls a callback for each row of the result set
	SelectWithPaging(page int, callback ProcessRow, query string, value ...interface{}) error

	// GetRandomUUID returns the cassandra uuid
	GetRandomUUID() (string, error)

	// Closed function to check is session is closed or not
	Closed() bool
}

// BatchDbConnector interface is responsible for dealing with the database using batches
type BatchDbConnector interface {
	// BatchExecution executing batch
	BatchExecution(query string, values [][]interface{}) error

	// Close method should be called by the caller at the end of all operation to Close the Db connection.
	Close()

	// Closed function to check is session is closed or not
	Closed() bool
}

// FullDbConnector interface is responsible for dealing with the database (with batches)
type FullDbConnector interface {
	DbConnector
	BatchDbConnector
}

const (
	//ErrDbHostsAndKeyspaceRequired error code for blank or no cluster hostnames and/or keyspace
	ErrDbHostsAndKeyspaceRequired = "DbHostsAndKeyspaceRequired"

	//ErrDbUnableToConnect error code for unable to connect to the DB with the given input
	ErrDbUnableToConnect = "DbUnableToConnect"

	//ErrDbDMLFailed error code for insert/update/delete failed
	ErrDbDMLFailed = "DbDMLFailed"

	//ErrDbNoOpenConnection error code for no open connection to connect to Cassandra
	ErrDbNoOpenConnection = "DbNoOpenConnection"

	//ErrDbUnableToFetchRecord error code for select query returned an error
	ErrDbUnableToFetchRecord = "DbUnableToFetchRecord"

	//ErrDbNoRecordToFetch error code for select query returned an error
	ErrDbNoRecordToFetch = "ErrDbNoRecordToFetch"

	//ErrDbIterClose error code denotes error while closing iterator
	ErrDbIterClose = "ErrDbIterClose"

	//ErrUUID error code for uuid generation
	ErrUUID = "ErrUUID"

	//ErrBatchExecution error code for failed batch execution
	ErrBatchExecution = "ErrBatchExecution"

	//ErrPolicyNotDefined error code for policy not defined
	ErrPolicyNotDefined = "ErrPolicyNotDefined"
)
