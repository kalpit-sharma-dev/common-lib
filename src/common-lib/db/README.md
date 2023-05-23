# Description
This is a common implementation for SQL data access layer.

### Third-Party Libraties
#### sqlx
  - **Name** : sqlx
  - **Link** : https://github.com/jmoiron/sqlx
  - **License** : [MIT License] (https://github.com/jmoiron/sqlx/blob/master/LICENSE)
  - **Description** : General purpose extensions to Golang's database/sql.

#### MSSQL Driver
  - **Name** : Go Mssql Driver
  - **Link** : https://github.com/denisenkom/go-mssqldb
  - **License** : [BSD 3-Clause "New" or "Revised" License] (https://github.com/denisenkom/go-mssqldb/blob/master/LICENSE.txt)
  - **Description** : Golang Microsoft SQL Server library.

#### PostgreSQL Driver
  - **Name** : Go PostgreSQL Driver
  - **Link** : https://github.com/lib/pq
  - **License** : [License] (https://github.com/lib/pq/blob/master/LICENSE.md)
  - **Description** : Golang PostgreSQL Server driver library.

#### Go Cache
  - **Name** : Go Cache
  - **Link** : https://github.com/patrickmn/go-cache
  - **License** : [MIT License] (https://github.com/patrickmn/go-cache/blob/master/LICENSE)
  - **Description** : go-cache is an in-memory key:value store/cache used for caching the prepared statements

#### Go-sqlmock
  - **Name** : go-sqlmock
  - **Link** : https://github.com/DATA-DOG/go-sqlmock
  - **License** : [BSD 3-Clause] (https://github.com/DATA-DOG/go-sqlmock/blob/master/LICENSE)
  - **Description** : go-sqlmock is a mock library implementing sql/driver.

### Note on Transactions

After starting a transaction and checking for an error in starting it, you should defer rolling it back. Example:

```go
tx, err := service.repo.BeginTransaction(ctx, partnerID)
if err != nil {
    return err
}
defer tx.Rollback()

foo, err = service.add(ctx, tx, fooInput)
if err != nil {
    return err
}

if err = tx.Commit(); err != nil {
    return err
}

return foo
```

If the transaction gets committed, then it cannot be rolled back, so calling `tx.Rollback()` won't cause any issues.

The context, `ctx`, passed into `BeginTransaction` can technically remove the need for deferring a rollback, because when the context ends, the transaction will be rolled back. However, if someone passes in a context that never ends, then we'd run into an issue where we'd leave transactions open indefinitely and run out of database connections (for example, when working with our Kafka consumers, we currently do not provide them with a context). In order to avoid mistakes like someone using a context that doesn't end properly, we should always defer the rollback of a transaction. While you may intend for your code to only be used in some sort of safe environment that always has proper contexts, like serving HTTP requests, someone else may one day call your code from a different environment, like a Kafka consumer, and not provide a good context (simply because they don't understand how important the cancellation is - for example, they may use `context.Background()`). In addition, if they were to make such a mistake, it may not be caught until too late.

### Use

#### Glide Dependencies
If glide is being used for dependency management, following dependency SHOULD be added to glide.yaml to avoid version mismatch errors during compilation and execution

```yaml
package: github.com/jmoiron/sqlx
version: 0794cb1f47ee444eda9624f952ab8a370bec22de
```
Apart from this, depending upon the database server in use the dependency for driver may also be added

**Import Statement**

```go
import (
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/db"

	//Import only one of the following two depending upon the SQL server
	//Import for loading mssql driver
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/db/mssql"
	//Import for loading postgresql driver
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/db/postgresql"
)
```

**Supported Drivers**
* mssql
* postgresql

**Configuration**

```go
//Config is struct to define db configurations
type Config struct {
	//DbName - Db to be selected after connecting to server
	//Required
	DbName string

	//Server - ip address of db host server
	//Required
	Server string

	//UserID - UserId for db server
	//Required
	UserID string

	//Password - Password for db server
	//Required
	Password string

	//Driver - Name of db driver
	//Required
	Driver string

	//Map to hold additional db config
	AdditionalConfig map[string]string

	//CacheLimit - CacheLimit sets limit on number of prepared statements to be cached
	//Default CacheLimit: 100
	CacheLimit int

	// CB struct contains configuration values for database circuit breaker.
	CB CircuitBreakerConfig
}
```
In case of PostgreSQL, following additional configuration (using the *AdditionalConfig* map in *Config*) may be provided
The keys to be used in the map are available in *postgresql* package

- Database Server Port : Key=postgresql.ServerPortKey, Value=Integer (as string) value of port
- Database SSL Mode : Key=postgresql.SSLModeKey, Value=Valid SSL Mode string

See example below for typical use

**Callback Functions**
```go
//ProcessObject is a callback function used to process one object which represents one table row.
//This is typically used when selecting large dataset which should be processed one row at a time.
//The callback function is called for each row read from db and mapped to structure object
//the object param holds the pointer to user defined structure which maps to row.
//the err param holds any error during the operation. In case of no error, it would be nil
//If this callback function returns error, the processing is stopped and no more calls to callback function occur
type ProcessObject func(object interface{}, err error) error
```
```go
//ProcessRow is a callback function used to process table row
type ProcessRow func(row Row)
```


**DatabaseProvider Instance**
```go
	//GetDbProvider - Fetching and initializing Database Provider using db configurations.
	//Once it is invoked it will return instance of provider struct to access db.
	//  Returns - DatabaseProvider : instance of provider struct to access db
	//error : incase db gives error creating connection for given configs
	db, err := db.GetDbProvider(db.Config{DbName: "NOCBO",
		Server:     "10.2.27.41",
		Password:   "its",
		UserID:     "its",
		Driver:     mssql.Dialect,
		CacheLimit: 200,
		// Cicuit breaker may or may not be enabled for a db.
		/* If enabled and config values are not provided, it will use default values as specified below.*/
		// It is recommended to set new config values for your db.
		CB: db.CircuitBreakerConfig{
			CircuitBreaker: &circuit.Config{
				Enabled:                true,
				TimeoutInSecond:        3,
				MaxConcurrentRequests:  15000,
				ErrorPercentThreshold:  25,
				RequestVolumeThreshold: 10,
 				SleepWindowInSecond:    10,
 			}}
		})
```
****

**Interface Functions**
```go
//DatabaseProvider is interface that holds all the functions related to Db
type DatabaseProvider interface {
	//GetSingleConnectionProvider returns a single connection from the provider
	//This takes a connection from the pool so you can run multiple requests on one connections
	//  Returns - DatabaseProvider : instance of provider struct to access db.
	//error: incase db provider gives error getting connection
	//  Note: Connection needs to be closed when done with it to return to the pool
	GetSingleConnectionProvider(ctx context.Context) (DatabaseConnectionProvider, error)

	//SelectWithPrepare is used to execute select query using prepared statment.
	//This will fetch the results on the basis of query provided and the values.
	//  Ex: SelectWithPrepare(someQuery,val1,val2,val3)
	//Returns - []map[string]interface{}: this will contain rows of data.
	//error: incase the database gets error creating prepared statement or query.
	//Note: Incase query returns large result data, all the data rows will be returned at once.
	//Note : Consider using SelectObjectsWithPrepare instead
	SelectWithPrepare(query string, value ...interface{}) ([]map[string]interface{}, error)

	//ExecWithPrepare is used to execute query that does not return data rows using prepared statement ex: INSERT, UPDATE or DELETE.
	//  Ex: ExecWithPrepare(someQuery,val1,val2,val3)
	//Returns - error: incase the database get error creating prepared statement or executing query.
	ExecWithPrepare(query string, value ...interface{}) error

	//Select is used to execute plaintext select query.
	//  Ex: Select(someQuery)
	//Returns - []map[string]interface{}: this contains rows of data.
	//error: incase the database gets error executing the query
	//Note: Incase query returns large result data, all the data rows will be returned at once.
	//Note : Consider using SelectObjects instead
	Select(query string) ([]map[string]interface{}, error)

	//Exec is used to execute plaintext query that does not return data rows ex: INSERT, UPDATE or DELETE.
	//  Ex: Exec(someQuery).
	//Returns - error: incase database gets error executing the query.
	Exec(query string) error

	//SelectAndProcess is used to execute plaintext select query.
	//SelectAndProcess will process each row returned by query using a callback function.
	//	Ex: SelectAndProcess(someQuery, callbackFunction)
	SelectAndProcess(query string, callback ProcessRow)

	//SelectWithPrepareAndProcess is used to execute select query using prepared statement.
	//SelectWithPrepareAndProcess will process each row returned by query using a callback function
	//  Ex: SelectWithPrepareAndProcess(someQuery, callbackFunction, val1,val2...)
	SelectWithPrepareAndProcess(query string, callback ProcessRow, value ...interface{})

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
	//Check the documentation of ProcessObject for details on callback function parameter and expected behaviour
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
	//Check the documentation of ProcessObject for details on callback function parameter and expected behaviour
	//param - transactionID : the application transaction ID of this DB operation. Typically used for logging
	//param - object : the pointer to object which shall be populated by each scanned row one by one
	//param - callback : the callback function called by this method upon getting next scanned row or getting error
	//param - query : the select prepared statment query string
	//param - value : the binding parameters associated with prepared statement
	//Returns - this method does not return any result as the result is populated in the objects parameter
	//error: incase of any error during operation, the callback function is called by passing the error to it
	SelectObjectWithPrepareAndProcess(transactionID string, object interface{}, callback ProcessObject, query string, value ...interface{})

	//PrepareStatement uses the provided query string to create a PreparedStatement in db server
	//The PreparedStatement prepared by this method is cached locally which can
	//be directly referred by various methods of this interface which take PreparedStatement as parameter
	PrepareStatement(transactionID string, query string) error

	//CloseStatement is used to close prepared statement created for given query.
	//Returns - error: if database get error closing prepared statement.
	CloseStatement(query string) error

	//BeginTransaction - starts a new transaction
	// Transaction can be started on db provider or single connection provider
	//Returns error if database fails to start a transaction
	BeginTransaction(ctx context.Context) (Transaction, error)
}

```

```go
//DatabaseConnectionProvider is inteface that holds all the functions related to a Db connection
//It holds all of the functionality of DatabaseProvider as well as connection specific functionality
type DatabaseConnectionProvider interface {
	DatabaseProvider

	//Close returns the connection back to the connection pool
	//All calls to the provider after calling Close will error
	//Returns - this method does not return any result but will instead mark the provider as closed
	//error: incase of any error during operation, the associated error is returned
	Close() error
}
```

```go
// Transaction holds all the functions needed to implement database transactions
// Transaction can be started on db provider or single connection provider
type Transaction interface {

	//ExecWithPrepareContext is used to execute query in transaction that does not return data rows - INSERT, UPDATE or DELETE.
	//	Note: Prepared statement is created internally for every query and not cached.
	//  Ex: ExecWithPrepareContext(ctx, someQuery,val1,val2,val3)
	//Returns - error: incase the database get error executing query.
	ExecWithPrepareContext(ctx context.Context, query string, value ...interface{}) error

	//ExecContext is used to execute plaintext query in transaction that does not return data rows - INSERT, UPDATE or DELETE.
	//  Ex: ExecContext(ctx, someQuery).
	//Returns - error: incase database gets error executing the query.
	ExecContext(ctx context.Context, query string) error

	//SelectObjectsWithPrepareContext is used to execute select query using prepared statement in transaction.
	//  Note: Prepared statement is created internally for every query and not cached.
	//This method unmarshalls the returned row(s) to slice of user defined structure object pointed by objects parameter
	//The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
	//It shall be noted that 'objects' must be of type pointer to slice of actual object even when single record is returned. The size of slice indicates number of records returned.
	//	Returns - this method does not return any result as the result is populated in the objects parameter
	//error: incase of any error during operation, the associated error is returned
	//Note: Incase query returns large result data, all the data rows will be returned at once.
	SelectObjectsWithPrepareContext(ctx context.Context, objects interface{}, query string, value ...interface{}) error

	//SelectObjectsContext is used to execute plaintext select query.
	//
	//This method unmarshalls the returned row(s) to slice of user defined structure object pointed by objects parameter.
	//The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
	//It shall be noted that 'objects' must be of type pointer to slice of actual object even when single record is returned. The size of slice indicates number of records returned.
	//	Returns - this method does not return any result as the result is populated in the objects parameter
	//error: incase of any error during operation, the associated error is returned
	//  Note: Incase query returns large result data, all the data rows will be returned at once.
	SelectObjectsContext(ctx context.Context, objects interface{}, query string) error

	//SelectObjectAndProcessContext is used to execute plaintext select query in transaction.
	//
	//The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
	//It shall be noted that 'object' must be of type pointer to actual object.
	//The method scans one row at a time and marshalls it to strcture object. It then calls the callback function of type ProcessObject.
	//Check the documentation of ProcessObject for details on callback function parameter and expected behaviour
	//	Returns - this method does not return any result as the result is populated in the objects parameter
	//error: incase of any error during operation, the callback function is called by passing the error to it
	SelectObjectAndProcessContext(ctx context.Context, object interface{}, callback ProcessObject, query string)

	//SelectObjectWithPrepareAndProcessContext is used to execute select query using prepared statement in transaction.
	//  Note: Prepared statement is created internally and not cached
	//The structure fields must be tagged with standard 'db' tag providing the column name associated with the field
	//It shall be noted that 'object' must be of type pointer to actual object.
	//The method scans one row at a time and marshalls it to strcture object. It then calls the callback function of type ProcessObject.
	//Check the documentation of ProcessObject for details on callback function parameter and expected behaviour
	//	Returns - this method does not return any result as the result is populated in the objects parameter
	//error: incase of any error during operation, the callback function is called by passing the error to it
	SelectObjectWithPrepareAndProcessContext(ctx context.Context, object interface{}, callback ProcessObject, query string, value ...interface{})

	// Commit commits the transaction.
	// Returns - error if transaction commit fails
	Commit() error

	// Rollback aborts the transaction.
	// Returns - error if transaction rollback fails
	Rollback() error
}
```

**Registering Dialect**
- To register a dialect set config.Driver as Dialect name

**Note** :
- Cache has limit on caching number of items (config.CacheLimit or defaultCacheLimit = 100). On exceeding this limit all the cache data will be flushed.

**Example**
	[Please refer:](gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/db/example/example.go)