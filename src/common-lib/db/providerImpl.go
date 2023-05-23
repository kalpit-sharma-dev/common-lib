package db

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	errs "github.com/pkg/errors"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

var (
	instance      *provider
	providerLock  sync.Mutex
	providerCache = make(map[string]DatabaseProvider)
)

type provider struct {
	driver     string
	datasource string
	db         *sqlx.DB
	config     Config
	dialect    dialect
}

// GetDbProvider - Fetching and initializing Database Provider using db configurations.
// Once it is invoked it will return instance of provider struct to access db.
//
//	Returns - DatabaseProvider : instance of provider struct to access db.
//
// error : incase db gives error creating connection for given configs.
func GetDbProvider(config Config) (DatabaseProvider, error) {
	if config.CircuitBreaker.Config != nil &&
		config.CircuitBreaker.Config.Enabled {
		circuit.Logger = logger.Get

		err := circuit.Register("", config.Server+"_"+config.DbName, config.CircuitBreaker.circuitBreaker(), config.CircuitBreaker.StateChangeCallback)
		if err != nil {
			return nil, errs.Wrapf(err, "Failed to register circuit breaker for db: %s", config.DbName)
		}
	} else {
		//nolint:exhaustivestruct
		config.CircuitBreaker.Config = &circuit.Config{
			Enabled: false,
		}
	}

	initializeCache(config)

	dialect, ok := getDialect(config.Driver)
	if !ok {
		return nil, fmt.Errorf("GetDbProvider: failed to get dialect instance for " + config.Driver)
	}
	dbConnInfo, err := dialect.GetConnectionString(config)
	if err != nil {
		return nil, fmt.Errorf("GetDbProvider: failed to get database connection config. " + err.Error())
	}

	providerLock.Lock()
	defer providerLock.Unlock()
	if providerInstance, ok := providerCache[dbConnInfo]; !ok {
		database, err := getConnection(config.Driver, dbConnInfo)
		if err != nil {
			return nil, err
		}

		instance = &provider{
			datasource: dbConnInfo,
			driver:     config.Driver,
			db:         database,
			config:     config,
			dialect:    dialect,
		}
		providerCache[dbConnInfo] = instance
	} else {
		instance = providerInstance.(*provider)
	}

	return instance, nil
}

// GetSingleConnectionProvider returns a single connection from the provider
// This takes a connection from the pool so you can run multiple requests on one connections
//
//	Returns - DatabaseProvider : instance of provider struct to access db.
//
// error: incase db provider gives error getting connection
//
//	Note: Either the context or the Connection needs to be closed when done with it to return to the pool
func (c *provider) GetSingleConnectionProvider(ctx context.Context, transactionID string) (DatabaseConnectionProvider, error) {
	var (
		err        error
		connection *sqlx.Conn
	)
	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		connection, err = c.db.Connx(ctx)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	if err != nil {
		return nil, fmt.Errorf("GetSingleConnectionProvider: failed to get the connection. " + err.Error())
	}

	newCtx, cancel := context.WithCancel(ctx)

	connectionProvider := &connProvider{
		conn:      connection,
		ctx:       newCtx,
		ctxCancel: cancel,
		stmts:     make(map[string]*sqlx.Stmt),
		closed:    make(chan struct{}),
		config:    c.config,
		dialect:   c.dialect,
	}

	go connectionProvider.monitorContext(transactionID)

	return connectionProvider, nil
}

// convertSQLRowsToMap returns rows from db as []map[string]interface{}
func convertSQLRowsToMap(rows *sqlx.Rows) ([]map[string]interface{}, error) {
	err := errors.New("ConvertSQLRowsToMap: Invalid sql rows")
	count := 0
	if rows == nil {
		return nil, err
	}
	rowsHolder := make([]map[string]interface{}, 0, 1)
	defer rows.Close() //nolint:errcheck
	for rows.Next() {
		count++
		cols, er := rows.Columns()
		if er != nil {
			return nil, er
		}
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		err = rows.Scan(columnPointers...)

		if err != nil {
			return nil, err
		}
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		rowsHolder = append(rowsHolder, m)
		err = nil
	}
	if count == 0 {
		return rowsHolder, nil
	}
	return rowsHolder, err
}

var getConnection = func(driver string, datasource string) (*sqlx.DB, error) {
	return sqlx.Connect(driver, datasource)
}

// ExecWithPrepare is used to execute query that does not return data rows using prepared statement ex: INSERT, UPDATE or DELETE.
//
//	Ex: ExecWithPrepare(someQuery,val1,val2,val3)
//
// Returns - error: incase the database get error creating prepared statement or executing query.
func (c *provider) ExecWithPrepare(query string, value ...interface{}) error {
	var (
		err  error
		stmt *sqlx.Stmt
	)

	stmt, err = c.prepareStatementInt("", query)
	if err != nil {
		return err
	}

	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		_, err = stmt.Exec(value...)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	return err
}

// SelectWithPrepare is used to execute select query using prepared statement.
// This will fetch the results on the basis of query provided and the values.
//
//	Ex: SelectWithPrepare(someQuery,val1,val2,val3)
//
// Returns - []map[string]interface{}: this will contain rows of data.
// error: incase the database gets error creating prepared statement or query execution.
//
//	Note: Incase query returns large result data, all the data rows will be returned at once.
func (c *provider) SelectWithPrepare(query string, value ...interface{}) ([]map[string]interface{}, error) {
	var (
		err  error
		stmt *sqlx.Stmt
		rows *sqlx.Rows
	)

	stmt, err = c.prepareStatementInt("", query)
	if err != nil {
		return nil, err
	}

	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		rows, err = stmt.Queryx(value...)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	return convertSQLRowsToMap(rows)
}

// Exec is used to execute plaintext query that does not return data rows ex: INSERT, UPDATE or DELETE.
//
//	Ex: Exec(someQuery).
//
// Returns - error: incase database gets error executing the query.
func (c *provider) Exec(query string) error {
	var err error

	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		_, err = c.db.Exec(query)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	return err
}

// Select is used to execute plaintext select query.
//
//	Ex: Select(someQuery)
//
// Returns - []map[string]interface{}: this contains rows of data.
// error: incase the database gets error executing the query
//
//	Note: Incase query returns large result data, all the data rows will be returned at once.
func (c *provider) Select(query string) ([]map[string]interface{}, error) {
	var (
		records []map[string]interface{}
		rows    *sqlx.Rows
		err     error
	)

	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		rows, err = c.db.Queryx(query)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	records, err = convertSQLRowsToMap(rows)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// CloseStatement is used to close prepared statement created for given query.
// Returns - error: if database get error closing prepared statement.
func (c *provider) CloseStatement(query string) error {
	var (
		err error
		st  *sqlx.Stmt
	)

	st = getStatement(query)
	if st == nil {
		return nil
	}

	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		err = st.Close()

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	if err != nil {
		return err
	}

	deleteKey(query)

	return nil
}

// PrepareStatement uses the provided query string to create a PreparedStatement in db server
// The PreparedStatement prepared by this method is cached locally which can
// be directly referred by various methods of this interface which take PreparedStatement as parameter
func (c *provider) PrepareStatement(transactionID string, query string) error {
	_, err := c.prepareStatementInt(transactionID, query)
	return err
}

// SelectAndProcess is used to execute plaintext select query.
// SelectAndProcess will process each row returned by query using a callback function.
//
//	Ex: SelectAndProcess(someQuery, callbackFunction)
func (c *provider) SelectAndProcess(query string, callback ProcessRow) {
	var (
		err  error
		rows *sqlx.Rows
	)

	//nolint:errcheck
	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		rows, err = c.db.Queryx(query)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	if err != nil {
		callback(Row{Columns: nil, Error: err})

		return
	}

	processRows(rows, callback)
}

// SelectWithPrepareAndProcess is used to execute select query using prepared statement.
// SelectWithPrepareAndProcess will process each row returned by query using a callback function
//
//	Ex: SelectWithPrepareAndProcess(someQuery, callbackFunction, val1,val2...)
func (c *provider) SelectWithPrepareAndProcess(query string, callback ProcessRow, value ...interface{}) {
	var (
		err  error
		stmt *sqlx.Stmt
		rows *sqlx.Rows
	)

	stmt, err = c.prepareStatementInt("", query)
	if err != nil {
		callback(Row{Columns: nil, Error: err})

		return
	}

	//nolint:errcheck
	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		rows, err = stmt.Queryx(value...)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	if err != nil {
		callback(Row{Columns: nil, Error: err})

		return
	}

	processRows(rows, callback)
}

func (c *provider) SelectObjectsWithPrepare(transactionID string, objects interface{}, query string, value ...interface{}) error {
	var (
		err  error
		stmt *sqlx.Stmt
	)

	stmt, err = c.prepareStatementInt(transactionID, query)
	if err != nil {
		return err
	}

	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		err = stmt.Select(objects, value...)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	return err
}

func (c *provider) SelectObjects(transactionID string, objects interface{}, query string) error {
	var err error

	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		err = c.db.Select(objects, query)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	return err
}

func (c *provider) SelectObjectAndProcess(transactionID string, object interface{}, callback ProcessObject, query string) {
	var (
		err  error
		rows *sqlx.Rows
	)

	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		rows, err = c.db.Queryx(query)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	if err != nil {
		//nolint:gosec
		//nolint:errcheck
		callback(object, err)

		return
	}

	processRowsForObjects(rows, object, callback)
}

func (c *provider) SelectObjectWithPrepareAndProcess(transactionID string, object interface{}, callback ProcessObject, query string, value ...interface{}) {
	var (
		rows *sqlx.Rows
		stmt *sqlx.Stmt
		err  error
	)

	stmt, err = c.prepareStatementInt(transactionID, query)
	if err != nil {
		//nolint:gosec
		callback(object, err)
		return
	}

	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		rows, err = stmt.Queryx(value...)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	if err != nil {
		//nolint:gosec
		//nolint:errcheck
		callback(object, err)

		return
	}

	processRowsForObjects(rows, object, callback)
}

func (c *provider) prepareStatementInt(transactionID string, query string) (*sqlx.Stmt, error) {
	var (
		stmt  *sqlx.Stmt
		err   error
		cbErr error
	)

	stmt = getStatement(query)
	if stmt == nil {
		Logger().Info(transactionID, "Creating new prepared statement")
		Logger().Debug(transactionID, "Prepared statement query: "+query)
		cbErr = circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
			stmt, err = c.db.Preparex(query)

			return c.dialect.ValidCbError(err)
		}, nil)
	}

	if cbErr != nil {
		err = cbErr
	}

	if err != nil {
		//nolint:goerr113
		//nolint:errorlint
		return nil, fmt.Errorf("Failed to prepare a statement. Error: %v", err)
	}

	addStatement(transactionID, query, stmt)
	Logger().Info(transactionID, "Prepared statement created")

	return stmt, nil
}

// processRows process row returned by query using callback function one row at time
func processRows(rows *sqlx.Rows, callback ProcessRow) {
	defer rows.Close() //nolint
	columnNames, err := rows.Columns()
	if err != nil {
		callback(Row{Error: fmt.Errorf("Failed to find column %+v", err)})
		return
	}

	colCount := len(columnNames)
	columnValues := make([]interface{}, colCount)
	columnPtrs := make([]interface{}, colCount)

	for rows.Next() {
		for i := 0; i < colCount; i++ {
			columnPtrs[i] = &columnValues[i]
		}
		err = rows.Scan(columnPtrs...)
		if err != nil {
			callback(Row{Error: err})
			continue
		}

		columns := make([]Column, colCount)
		for i := 0; i < colCount; i++ {
			columns[i] = Column{Name: columnNames[i], Value: columnValues[i]}
		}
		callback(Row{Columns: columns})
	}
}

func processRowsForObjects(rows *sqlx.Rows, object interface{}, callback ProcessObject) {
	defer rows.Close() //nolint
	for rows.Next() {
		err := rows.StructScan(object)
		if e := callback(object, err); e != nil {
			return
		}
	}
	if err := rows.Err(); err != nil {
		callback(object, err)
	}
}

// BeginTransaction - starts a new transaction.
//
//	Returns error if database fails to start a transaction
func (c *provider) BeginTransaction(ctx context.Context) (Transaction, error) {
	var (
		tx  *sqlx.Tx
		err error
	)

	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		tx, err = c.db.BeginTxx(ctx, nil)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	return &dbTx{tx, &c.config, c.dialect}, nil
}
