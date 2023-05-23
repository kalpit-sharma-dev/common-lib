package db

import (
	"context"
	"errors"
	"sync"

	"github.com/jmoiron/sqlx"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

type connProvider struct {
	conn      *sqlx.Conn
	ctx       context.Context
	ctxCancel context.CancelFunc
	stmts     map[string]*sqlx.Stmt
	mu        sync.RWMutex
	closed    chan struct{}
	config    Config
	dialect   dialect
}

// GetSingleConnectionProvider returns a single connection from the provider
// This takes a connection from the pool so you can run multiple requests on one connections
//
//	Returns - DatabaseProvider : instance of provider struct to access db.
//
// error: incase db provider gives error getting connection
//
//	Note: Connection needs to be closed when done with it to return to the pool
func (c *connProvider) GetSingleConnectionProvider(ctx context.Context, transactionID string) (DatabaseConnectionProvider, error) {
	return c, nil
}

// Close returns the connection back to the connection pool. Will also close all statements associated to this connProvider
// All calls to the provider after calling Close will error
// Returns - this method does not return any result but will instead mark the provider as closed
// error: incase of any error during operation, the associated error is returned
//
//	Note: Closing the context calls this behind the scene if not alraedy closed.
func (c *connProvider) Close(transactionID string) error {
	c.mu.Lock()

	if c.closed != nil {
		// Sending a message to the channel to close the monitor goroutine
		c.closed <- struct{}{}
		// Closing the channel so future calls will be no-ops
		c.closed = nil
	} else {
		// Provider already closed to returning early
		c.mu.Unlock()
		return nil
	}

	if c.ctx.Err() == nil {
		c.ctxCancel()
	}

	for _, stmt := range c.stmts {
		stmt.Close()
	}
	c.stmts = nil
	c.mu.Unlock()

	err := c.conn.Close()
	return err
}

func (c *connProvider) monitorContext(transactionID string) {
	for {
		select {
		case <-c.closed:
			// connection closed so return to exit the loop
			return
		case <-c.ctx.Done():
			Logger().Info(transactionID, "Context Closed so closing connection")
			// doing the following call in a seperate goroutine so it doesnt block the listener and doesnt cause deadlocks
			go c.Close(transactionID)
		}
	}
}

// ExecWithPrepare is used to execute query that does not return data rows using prepared statement ex: INSERT, UPDATE or DELETE.
//
//	Ex: ExecWithPrepare(someQuery,val1,val2,val3)
//
// Returns - error: incase the database get error creating prepared statement or executing query.
func (c *connProvider) ExecWithPrepare(query string, value ...interface{}) error {
	stmt, err := c.prepareStatementInt("", query)
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
func (c *connProvider) SelectWithPrepare(query string, value ...interface{}) ([]map[string]interface{}, error) {
	var (
		err  error
		rows *sqlx.Rows
	)

	stmt, err := c.prepareStatementInt("", query)
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
		return nil, err
	}
	records, err := convertSQLRowsToMap(rows)
	if err != nil {
		return nil, err
	}
	return records, nil
}

// Exec is used to execute plaintext query that does not return data rows ex: INSERT, UPDATE or DELETE.
//
//	Ex: Exec(someQuery).
//
// Returns - error: incase database gets error executing the query.
func (c *connProvider) Exec(query string) error {
	var err error

	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		_, err := c.conn.ExecContext(c.ctx, query)

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
func (c *connProvider) Select(query string) ([]map[string]interface{}, error) {
	var (
		err  error
		rows *sqlx.Rows
	)
	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		rows, err = c.conn.QueryxContext(c.ctx, query)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	if err != nil {
		return nil, err
	}
	records, err := convertSQLRowsToMap(rows)
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CloseStatement is used to close prepared statement created for given query for the connProvider.
// Returns - error: if database get error closing prepared statement.
func (c *connProvider) CloseStatement(query string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	st := c.stmts[query]

	if st == nil {
		return nil
	}
	err := st.Close()
	if err != nil {
		return err
	}
	delete(c.stmts, query)
	return nil
}

// PrepareStatement uses the provided query string to create a PreparedStatement for the connProvider
// The PreparedStatement prepared by this method is cached locally which can
// be directly referred by various methods of this interface which take PreparedStatement as parameter
func (c *connProvider) PrepareStatement(transactionID string, query string) error {
	_, err := c.prepareStatementInt(transactionID, query)
	return err
}

// SelectAndProcess is used to execute plaintext select query.
// SelectAndProcess will process each row returned by query using a callback function.
//
//	Ex: SelectAndProcess(someQuery, callbackFunction)
func (c *connProvider) SelectAndProcess(query string, callback ProcessRow) {
	var (
		err  error
		rows *sqlx.Rows
	)
	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		rows, err = c.conn.QueryxContext(c.ctx, query)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	if err != nil {
		callback(Row{Error: err})
		return
	}

	processRows(rows, callback)
}

// SelectWithPrepareAndProcess is used to execute select query using prepared statement.
// SelectWithPrepareAndProcess will process each row returned by query using a callback function
//
//	Ex: SelectWithPrepareAndProcess(someQuery, callbackFunction, val1,val2...)
func (c *connProvider) SelectWithPrepareAndProcess(query string, callback ProcessRow, value ...interface{}) {
	var (
		err  error
		rows *sqlx.Rows
	)
	stmt, err := c.prepareStatementInt("", query)
	if err != nil {
		callback(Row{Error: err})
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
		callback(Row{Error: err})
		return
	}

	processRows(rows, callback)
}

func (c *connProvider) SelectObjectsWithPrepare(transactionID string, objects interface{}, query string, value ...interface{}) error {
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

func (c *connProvider) SelectObjects(transactionID string, objects interface{}, query string) error {
	return c.conn.SelectContext(c.ctx, objects, query)
}

func (c *connProvider) SelectObjectAndProcess(transactionID string, object interface{}, callback ProcessObject, query string) {
	var (
		err  error
		rows *sqlx.Rows
	)
	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		rows, err = c.conn.QueryxContext(c.ctx, query)

		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	if err != nil {
		callback(object, err)
		return
	}
	processRowsForObjects(rows, object, callback)
}

func (c *connProvider) SelectObjectWithPrepareAndProcess(transactionID string, object interface{}, callback ProcessObject, query string, value ...interface{}) {
	var (
		err  error
		rows *sqlx.Rows
	)

	stmt, err := c.prepareStatementInt(transactionID, query)
	if err != nil {
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
		callback(object, err)
		return
	}
	processRowsForObjects(rows, object, callback)
}

func (c *connProvider) prepareStatementInt(transactionID string, query string) (*sqlx.Stmt, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	stmt := c.stmts[query]
	if stmt == nil {
		Logger().Info(transactionID, "Creating new prepared statement on this connection")
		Logger().Debug(transactionID, "Prepared statement query: "+query)
		st, err := c.conn.PreparexContext(c.ctx, query)
		if err != nil {
			return nil, err
		}
		if st == nil {
			return nil, errors.New("creating PreparedStatement failed")
		}
		c.stmts[query] = st
		Logger().Info(transactionID, "Prepared statement created")
		return st, nil
	}
	return stmt, nil
}

func (c *connProvider) BeginTransaction(ctx context.Context) (Transaction, error) {
	var (
		err error
		tx  *sqlx.Tx
	)
	cbErr := circuit.Do(c.config.Server+"_"+c.config.DbName, c.config.CircuitBreaker.Config.Enabled, func() error {
		tx, err = c.conn.BeginTxx(ctx, nil)
		return c.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}

	return &dbTx{tx, &c.config, c.dialect}, nil
}
