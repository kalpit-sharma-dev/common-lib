package db

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

type dbTx struct {
	tx       *sqlx.Tx
	dbconfig *Config
	dialect  dialect
}

// ExecWithPrepareContext is used to execute query in transaction that does not return data rows - INSERT, UPDATE or DELETE.
//
//		Note: Prepared statement is created internally for every query and not cached.
//	 Ex: ExecWithPrepareContext(ctx, someQuery,val1,val2,val3)
//
// Returns - error: incase the database get error executing query.
func (db *dbTx) ExecWithPrepareContext(ctx context.Context, query string, value ...interface{}) error {
	var err error
	cbErr := circuit.Do(db.dbconfig.Server+"_"+db.dbconfig.DbName, db.dbconfig.CircuitBreaker.Config.Enabled, func() error {
		_, err = db.tx.ExecContext(ctx, query, value...)

		return db.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	//nolint:wrapcheck
	return err
}

// ExecContext is used to execute plaintext query in transaction that does not return data rows - INSERT, UPDATE or DELETE.
//
//	Ex: ExecContext(ctx, someQuery).
//
// Returns - error: incase database gets error executing the query.
func (db *dbTx) ExecContext(ctx context.Context, query string) error {
	var err error
	cbErr := circuit.Do(db.dbconfig.Server+"_"+db.dbconfig.DbName, db.dbconfig.CircuitBreaker.Config.Enabled, func() error {

		_, err = db.tx.ExecContext(ctx, query)

		return db.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	//nolint:wrapcheck
	return err
}

// SelectObjectsWithPrepareContext is used to execute select query using prepared statement in transaction.
//
//	Note: Prepared statement is created internally for every query and not cached.
//
// This method unmarshalls the returned row(s) to slice of user defined structure object pointed by objects parameter
// The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
// It shall be noted that 'objects' must be of type pointer to slice of actual object even when single record is returned.
// The size of slice indicates number of records returned.
//
//	Returns - this method does not return any result as the result is populated in the objects parameter
//
// error: incase of any error during operation, the associated error is returned
// Note: Incase query returns large result data, all the data rows will be returned at once.
func (db *dbTx) SelectObjectsWithPrepareContext(ctx context.Context, objects interface{}, query string, value ...interface{}) error {
	var err error
	cbErr := circuit.Do(db.dbconfig.Server+"_"+db.dbconfig.DbName, db.dbconfig.CircuitBreaker.Config.Enabled, func() error {

		err = db.tx.SelectContext(ctx, objects, query, value...)

		return db.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	//nolint:wrapcheck
	return err
}

// SelectObjectsContext is used to execute plaintext select query.
//
// This method unmarshalls the returned row(s) to slice of user defined structure object pointed by objects parameter.
// The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
// It shall be noted that 'objects' must be of type pointer to slice of actual object even when single record is returned.
// The size of slice indicates number of records returned.
//
//	Returns - this method does not return any result as the result is populated in the objects parameter
//
// error: incase of any error during operation, the associated error is returned
//
//	Note: Incase query returns large result data, all the data rows will be returned at once.
func (db *dbTx) SelectObjectsContext(ctx context.Context, objects interface{}, query string) error {
	var err error
	cbErr := circuit.Do(db.dbconfig.Server+"_"+db.dbconfig.DbName, db.dbconfig.CircuitBreaker.Config.Enabled, func() error {

		err = db.tx.SelectContext(ctx, objects, query)

		return db.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	//nolint:wrapcheck
	return err
}

// SelectObjectAndProcessContext is used to execute plaintext select query in transaction.
//
// The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
// It shall be noted that 'object' must be of type pointer to actual object.
// The method scans one row at a time and marshalls it to strcture object. It then calls the callback function of type ProcessObject.
// Check the documentation of ProcessObject for details on callback function parameter and expected behavior.
//
//	Returns - this method does not return any result as the result is populated in the objects parameter
//
// error: incase of any error during operation, the callback function is called by passing the error to it
func (db *dbTx) SelectObjectAndProcessContext(ctx context.Context, object interface{}, callback ProcessObject, query string) {
	var (
		err  error
		rows *sqlx.Rows
	)
	cbErr := circuit.Do(db.dbconfig.Server+"_"+db.dbconfig.DbName, db.dbconfig.CircuitBreaker.Config.Enabled, func() error {

		rows, err = db.tx.QueryxContext(ctx, query)

		return db.dialect.ValidCbError(err)
	}, nil)
	if cbErr != nil {
		err = cbErr
	}
	if err != nil {
		if e := callback(object, err); e != nil {
			Logger().Error("", "Callback Error", "SelectObjectAndProcessContext - callback execution returned err: %v", e)

			return
		}

		return

	}
	processRowsForObjects(rows, object, callback)

}

// SelectObjectWithPrepareAndProcessContext is used to execute select query using prepared statement in transaction.
//
//	Note: Prepared statement is created internally and not cached.
//
// The structure fields must be tagged with standard 'db' tag providing the column name associated with the field.
// It shall be noted that 'object' must be of type pointer to actual object.
// The method scans one row at a time and marshalls it to strcture object. It then calls the callback function of type ProcessObject.
// Check the documentation of ProcessObject for details on callback function parameter and expected behavior.
//
//	Returns - this method does not return any result as the result is populated in the objects parameter.
//
// error: incase of any error during operation, the callback function is called by passing the error to it.
func (db *dbTx) SelectObjectWithPrepareAndProcessContext(ctx context.Context, object interface{},
	callback ProcessObject, query string, value ...interface{}) {
	var (
		err  error
		rows *sqlx.Rows
	)
	cbErr := circuit.Do(db.dbconfig.Server+"_"+db.dbconfig.DbName, db.dbconfig.CircuitBreaker.Config.Enabled, func() error {

		rows, err = db.tx.QueryxContext(ctx, query, value...)
		return db.dialect.ValidCbError(err)
	}, nil)
	if cbErr != nil {
		err = cbErr
	}
	if err != nil {
		if e := callback(object, err); e != nil {
			Logger().Error("", "Callback Error", "SelectObjectWithPrepareAndProcessContext - callback execution returned err: %v", e)

			return
		}

		return
	}

	processRowsForObjects(rows, object, callback)
}

// Commit commits the transaction.
// Returns - error if transaction commit fails.
func (db *dbTx) Commit() error {

	var err error
	cbErr := circuit.Do(db.dbconfig.Server+"_"+db.dbconfig.DbName, db.dbconfig.CircuitBreaker.Config.Enabled, func() error {

		err = db.tx.Commit()

		return db.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	//nolint:wrapcheck
	return err
}

// Rollback aborts the transaction.
// Returns - error if transaction rollback fails.
func (db *dbTx) Rollback() error {

	var err error
	cbErr := circuit.Do(db.dbconfig.Server+"_"+db.dbconfig.DbName, db.dbconfig.CircuitBreaker.Config.Enabled, func() error {

		err = db.tx.Rollback()

		return db.dialect.ValidCbError(err)
	}, nil)

	if cbErr != nil {
		err = cbErr
	}
	//nolint:wrapcheck
	return err
}
