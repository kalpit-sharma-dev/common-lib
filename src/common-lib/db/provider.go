// Package db provide implementation of SQL data access layer
package db

import "context"

//go:generate mockgen -package mock -destination=mock/mocks.go . DatabaseProvider,DatabaseConnectionProvider
//go:generate mockgen -package mock -destination=mock/mock_transaction.go . Transaction

//Row is struct to hold single table row
type Row struct {
	Columns []Column
	Error   error
}

//Column is a struct to hold all the column values of row
type Column struct {
	Name  string
	Value interface{}
}

//ProcessRow is a callback function used to process table row
type ProcessRow func(row Row)

//ProcessObject is a callback function used to process one object which represents one table row.
//This is typically used when selecting large dataset which should be processed one row at a time.
//The callback function is called for each row read from db and mapped to structure object
//the object param holds the pointer to user defined structure which maps to row.
//the err param holds any error during the operation. In case of no error, it would be nil
//If this callback function returns error, the processing is stopped and no more calls to callback function occur
type ProcessObject func(object interface{}, err error) error

//DatabaseProvider is interface that holds all the functions related to Db
type DatabaseProvider interface {
	//GetSingleConnectionProvider returns a single connection from the provider
	//This takes a connection from the pool so you can run multiple requests on one connections
	//  Returns - DatabaseProvider : instance of provider struct to access db.
	//error: incase db provider gives error getting connection
	//  Note: Connection needs to be closed when done with it to return to the pool
	GetSingleConnectionProvider(ctx context.Context, transactionID string) (DatabaseConnectionProvider, error)

	//SelectWithPrepare is used to execute select query using prepared statement.
	//This will fetch the results on the basis of query provided and the values.
	//  Ex: SelectWithPrepare(someQuery,val1,val2,val3)
	//Returns - []map[string]interface{}: this will contain rows of data.
	//error: incase the database gets error creating prepared statement or query execution.
	//  Note: Incase query returns large result data, all the data rows will be returned at once.
	//	Note : Consider using SelectObjectsWithPrepare instead of this method
	SelectWithPrepare(query string, value ...interface{}) ([]map[string]interface{}, error)

	//ExecWithPrepare is used to execute query that does not return data rows using prepared statement ex: INSERT, UPDATE or DELETE.
	//  Ex: ExecWithPrepare(someQuery,val1,val2,val3)
	//Returns - error: incase the database get error creating prepared statement or executing query.
	ExecWithPrepare(query string, value ...interface{}) error

	//Select is used to execute plaintext select query.
	//  Ex: Select(someQuery)
	//Returns - []map[string]interface{}: this contains rows of data.
	//error: incase the database gets error executing the query
	//  Note: Incase query returns large result data, all the data rows will be returned at once.
	//	Note : Consider using SelectObjects instead of this method
	Select(query string) ([]map[string]interface{}, error)

	//Exec is used to execute plaintext query that does not return data rows ex: INSERT, UPDATE or DELETE.
	//  Ex: Exec(someQuery).
	//Returns - error: incase database gets error executing the query.
	Exec(query string) error

	//SelectAndProcess is used to execute plaintext select query.
	//SelectAndProcess will process each row returned by query using a callback function.
	//	Ex: SelectAndProcess(someQuery, callbackFunction)
	// 	Note : Consider using SelectObjectAndProcess instead of this method
	SelectAndProcess(query string, callback ProcessRow)

	//SelectWithPrepareAndProcess is used to execute select query using prepared statement.
	//SelectWithPrepareAndProcess will process each row returned by query using a callback function
	//  Ex: SelectWithPrepareAndProcess(someQuery, callbackFunction, val1,val2...)
	// 	Note : Consider using SelectObjectWithPrepareAndProcess instead of this method
	SelectWithPrepareAndProcess(query string, callback ProcessRow, value ...interface{})

	//PrepareStatement uses the provided query string to create a PreparedStatement in db server
	//The PreparedStatement prepared by this method is cached locally which can
	//be directly referred by various methods of this interface which take PreparedStatement as parameter
	PrepareStatement(transactionID string, query string) error

	//CloseStatement is used to close prepared statement created for given query.
	//Returns - error: if database get error closing prepared statement.
	CloseStatement(query string) error

	//SelectObjectsWithPrepare behaves similar to SelectWithPrepare method of this interface, but apart from getting the data, it unmarshalls the returned row(s) to slice of user defined structure object pointed by objects parameter
	//The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
	//It shall be noted that 'objects' must be of type pointer to slice of actual object even when single record is returned. The size of slice indicates number of records returned.
	//param - transactionID : the application transaction ID of this DB operation. Typically used for logging
	//param - objects : the pointer to slice of objects which shall be populated by selected rows
	//param - query : the select prepared statment query string
	//param - value : the binding parameters associated with prepared statement
	//Returns - this method does not return any result as the result is populated in the objects parameter
	//error: incase of any error during operation, the associated error is returned
	//Note: Incase query returns large result data, all the data rows will be returned at once.
	SelectObjectsWithPrepare(transactionID string, objects interface{}, query string, value ...interface{}) error

	//SelectObjects behaves similar to 'Select' method of this interface, but apart from getting the data, it unmarshalls the returned row(s) to slice of user defined structure object pointed by objects parameter
	//The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
	//It shall be noted that 'objects' parameter must be of type pointer to slice of actual object even when single record is returned. The size of slice indicates number of records returned.
	//param - transactionID : the application transaction ID of this DB operation. Typically used for logging
	//param - objects : the pointer to slice of objects which shall be populated by selected rows
	//param - query : the actual select query string
	//Returns - this method does not return any result as the result is populated in the objects parameter
	//error: incase of any error during operation, the associated error is returned
	//Note: Incase query returns large result data, all the data rows will be returned at once.
	SelectObjects(transactionID string, objects interface{}, query string) error

	//SelectObjectAndProcess behaves similar to SelectAndProcess method of this interface, but apart from getting the data, it unmarshalls the returned row to user defined structure object pointed by object parameter
	//The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
	//It shall be noted that 'object' must be of type pointer to actual object.
	//The method scans one row at a time and marshalls it to strcture object. It then calls the callback function of type ProcessObject.
	//Check the documentation of ProcessObject for details on callback function parameter and expected behavior
	//param - transactionID : the application transaction ID of this DB operation. Typically used for logging
	//param - object : the pointer to object which shall be populated by each scanned row one by one
	//param - callback : the callback function called by this method upon getting next scanned row or getting error
	//param - query : the select query string
	//Returns - this method does not return any result as the result is populated in the objects parameter
	//error: incase of any error during operation, the callback function is called by passing the error to it
	SelectObjectAndProcess(transactionID string, object interface{}, callback ProcessObject, query string)

	//SelectObjectsWithPrepareAndProcess behaves similar to SelectWithPrepareAndProcess method of this interface, but apart from getting the data, it unmarshalls the returned row to user defined structure object pointed by object parameter
	//The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
	//It shall be noted that 'object' must be of type pointer to actual object.
	//The method scans one row at a time and marshalls it to strcture object. It then calls the callback function of type ProcessObject.
	//Check the documentation of ProcessObject for details on callback function parameter and expected behavior
	//param - transactionID : the application transaction ID of this DB operation. Typically used for logging
	//param - object : the pointer to object which shall be populated by each scanned row one by one
	//param - callback : the callback function called by this method upon getting next scanned row or getting error
	//param - query : the select prepared statment query string
	//param - value : the binding parameters associated with prepared statement
	//Returns - this method does not return any result as the result is populated in the objects parameter
	//error: incase of any error during operation, the callback function is called by passing the error to it
	SelectObjectWithPrepareAndProcess(transactionID string, object interface{}, callback ProcessObject, query string, value ...interface{})

	// BeginTransaction - starts a new transaction
	// Transaction can be started on db provider or single connection provider
	// Returns error if database fails to start a transaction
	BeginTransaction(ctx context.Context) (Transaction, error)
}

// DatabaseConnectionProvider is inteface that holds all the functions related to a Db connection
// It holds all of the functionality of DatabaseProvider as well as connection specific functionality.
type DatabaseConnectionProvider interface {
	DatabaseProvider

	// Close returns the connection back to the connection pool
	// All calls to the provider after calling Close will error
	// Returns - this method does not return any result but will instead mark the provider as closed
	// error: incase of any error during operation, the associated error is returned
	Close(transactionID string) error
}

// Transaction holds all the methods needed to implement database transactions
// Transaction can be started on db provider or single connection provider.
type Transaction interface {

	// ExecWithPrepareContext is used to execute query in transaction that does not return data rows - INSERT, UPDATE or DELETE.
	//	Note: Prepared statement is created internally for every query and not cached.
	//  Ex: ExecWithPrepareContext(ctx, someQuery,val1,val2,val3)
	// Returns - error: incase the database get error executing query.
	ExecWithPrepareContext(ctx context.Context, query string, value ...interface{}) error

	// ExecContext is used to execute plaintext query in transaction that does not return data rows - INSERT, UPDATE or DELETE.
	//  Ex: ExecContext(ctx, someQuery).
	// Returns - error: incase database gets error executing the query.
	ExecContext(ctx context.Context, query string) error

	// SelectObjectsWithPrepareContext is used to execute select query using prepared statement in transaction.
	//  Note: Prepared statement is created internally for every query and not cached.
	// This method unmarshalls the returned row(s) to slice of user defined structure object pointed by objects parameter
	// The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
	// It shall be noted that 'objects' must be of type pointer to slice of actual object even when single record is returned.
	// The size of slice indicates number of records returned.
	//	Returns - this method does not return any result as the result is populated in the objects parameter
	// error: incase of any error during operation, the associated error is returned
	// Note: Incase query returns large result data, all the data rows will be returned at once.
	SelectObjectsWithPrepareContext(ctx context.Context, objects interface{}, query string, value ...interface{}) error

	// SelectObjectsContext is used to execute plaintext select query.
	//
	// This method unmarshalls the returned row(s) to slice of user defined structure object pointed by objects parameter.
	// The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
	// It shall be noted that 'objects' must be of type pointer to slice of actual object even when single record is returned.
	// The size of slice indicates number of records returned.
	//	Returns - this method does not return any result as the result is populated in the objects parameter
	// error: incase of any error during operation, the associated error is returned
	//  Note: Incase query returns large result data, all the data rows will be returned at once.
	SelectObjectsContext(ctx context.Context, objects interface{}, query string) error

	// SelectObjectAndProcessContext is used to execute plaintext select query in transaction.
	//
	// The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
	// It shall be noted that 'object' must be of type pointer to actual object.
	// The method scans one row at a time and marshalls it to strcture object. It then calls the callback function of type ProcessObject.
	// Check the documentation of ProcessObject for details on callback function parameter and expected behavior
	//	Returns - this method does not return any result as the result is populated in the objects parameter
	// error: incase of any error during operation, the callback function is called by passing the error to it
	SelectObjectAndProcessContext(ctx context.Context, object interface{}, callback ProcessObject, query string)

	// SelectObjectWithPrepareAndProcessContext is used to execute select query using prepared statement in transaction.
	//  Note: Prepared statement is created internally and not cached
	// The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
	// It shall be noted that 'object' must be of type pointer to actual object.
	// The method scans one row at a time and marshalls it to strcture object. It then calls the callback function of type ProcessObject.
	// Check the documentation of ProcessObject for details on callback function parameter and expected behavior
	//	Returns - this method does not return any result as the result is populated in the objects parameter
	// error: incase of any error during operation, the callback function is called by passing the error to it
	SelectObjectWithPrepareAndProcessContext(ctx context.Context, object interface{}, callback ProcessObject, query string, value ...interface{})

	// Commit commits the transaction.
	// Returns - error if transaction commit fails
	Commit() error

	// Rollback aborts the transaction.
	// Returns - error if transaction rollback fails
	Rollback() error
}
