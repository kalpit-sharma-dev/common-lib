package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

const cbError string = "Valid circuit breaker error"

var postgresErrors = []string{
	"connection_exception", "connection_does_not_exist",
	"connection_failure", "bad connection", "connection refused",
	"sqlclient_unable_to_establish_sqlconnection",
	"sqlserver_rejected_establishment_of_sqlconnection", "broken pipe", "sql: connection is already closed",
}

type mockStruct struct{}

func (s mockStruct) GetConnectionString(config Config) (string, error) {
	if config.DbName == "" || config.Password == "" || config.Server == "" || config.UserID == "" {
		return "", fmt.Errorf("getDbConnInfo: One or more required db configuration  missing")
	}
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s", config.Server, config.UserID, config.Password, config.DbName)
	return connString, nil
}

func (s mockStruct) ValidCbError(err error) error {
	if err == nil {
		return nil
	}

	for _, v := range postgresqlErrors {
		if strings.Contains(err.Error(), v) {
			//nolint:goerr113
			return fmt.Errorf("%s : %v", cbError, err)
		}
	}

	return err
}

var isError bool

// Callback func to read rows
func process(row Row) {
	if row.Error != nil {
		isError = true
	}
}

func TestGetDbProvider(t *testing.T) {
	dialectsMap["mockMssql"] = mockStruct{}

	t.Run("Error failed to get dialect", func(t *testing.T) {
		_, err := GetDbProvider(Config{
			DbName:     "NOCBO",
			Server:     "10.2.27.41",
			Password:   "its",
			UserID:     "its",
			CacheLimit: 200,
		})

		if err == nil {
			t.Errorf("Expecting error but found nil")
		}
	})

	t.Run("Error to get connection Config", func(t *testing.T) {
		_, err := GetDbProvider(Config{
			Driver: "mockMssql",
		})
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}
	})

	t.Run("Error to get connection Config", func(t *testing.T) {
		old := getConnection
		getConnection = func(driver string, datasource string) (*sqlx.DB, error) {
			return nil, errors.New("Error getting db connection")
		}
		defer func() {
			getConnection = old
		}()
		_, err := GetDbProvider(Config{
			DbName:     "NOCBO",
			Server:     "10.2.27.41",
			Password:   "its",
			UserID:     "its",
			Driver:     "mockMssql",
			CacheLimit: 200,
		})
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}
	})

	t.Run("Success getting db provider when circuit breaker is not enabled", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer mockDB.Close()
		db := sqlx.NewDb(mockDB, "sqlmock")

		expectedProvider := &provider{
			driver:     "mockMssql",
			datasource: "server=10.2.27.41;user id=its;password=its;database=NOCBO",
			db:         db,
			dialect:    mockStruct{},
			config: Config{
				DbName:     "NOCBO",
				Server:     "10.2.27.41",
				Password:   "its",
				UserID:     "its",
				Driver:     "mockMssql",
				CacheLimit: 200,
				CircuitBreaker: CircuitBreaker{
					Config: &circuit.Config{
						Enabled: false,
					},
				},
			},
		}

		old := getConnection
		getConnection = func(driver string, datasource string) (*sqlx.DB, error) {
			return db, nil
		}
		defer func() {
			getConnection = old
		}()
		gotPrivder, er := GetDbProvider(Config{
			DbName:     "NOCBO",
			Server:     "10.2.27.41",
			Password:   "its",
			UserID:     "its",
			Driver:     "mockMssql",
			CacheLimit: 200,
		})
		if er != nil {
			t.Errorf("Expecting no error but found err %v", err)
		}

		if !reflect.DeepEqual(expectedProvider, gotPrivder) {
			t.Errorf("Expected provider %v but got provider %v", expectedProvider, gotPrivder)
		}
	})

	t.Run("Success getting db provider when circuit breaker is enabled", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer mockDB.Close()
		db := sqlx.NewDb(mockDB, "sqlmock")

		expectedProvider := &provider{
			driver:     "mockMssql",
			datasource: "server=10.2.27.42;user id=its;password=its;database=NOCBO_DB",
			db:         db,
			dialect:    mockStruct{},
			config: Config{
				DbName:     "NOCBO_DB",
				Server:     "10.2.27.42",
				Password:   "its",
				UserID:     "its",
				Driver:     "mockMssql",
				CacheLimit: 200,
				CircuitBreaker: CircuitBreaker{
					Config: &circuit.Config{
						Enabled:                true,
						TimeoutInSecond:        3,
						MaxConcurrentRequests:  15,
						ErrorPercentThreshold:  25,
						RequestVolumeThreshold: 50,
						SleepWindowInSecond:    10,
					},
				},
			},
		}

		old := getConnection
		getConnection = func(driver string, datasource string) (*sqlx.DB, error) {
			return db, nil
		}
		defer func() {
			getConnection = old
		}()
		gotPrivder, er := GetDbProvider(Config{
			DbName:     "NOCBO_DB",
			Server:     "10.2.27.42",
			Password:   "its",
			UserID:     "its",
			Driver:     "mockMssql",
			CacheLimit: 200,
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled:                true,
					TimeoutInSecond:        3,
					MaxConcurrentRequests:  15,
					ErrorPercentThreshold:  25,
					RequestVolumeThreshold: 50,
					SleepWindowInSecond:    10,
				},
			},
		})

		if er != nil {
			t.Errorf("Expecting no error but found err %v", err)
		}

		if !reflect.DeepEqual(expectedProvider, gotPrivder) {
			t.Errorf("Expected provider %v but got provider %v", expectedProvider, gotPrivder)
		}
	})
}

func TestGetSingleConnectionProvider(t *testing.T) {
	t.Run("Success getting single connection provider", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer mockDB.Close()
		db := sqlx.NewDb(mockDB, "sqlmock")

		p := provider{
			driver:     "mssql",
			datasource: "connection string",
			db:         db,
			dialect:    mockStruct{},
			config: Config{
				CircuitBreaker: CircuitBreaker{
					Config: &circuit.Config{
						Enabled: false,
					},
				},
			},
		}

		ctx := context.Background()

		_, err = p.GetSingleConnectionProvider(ctx, "")

		if err != nil {
			t.Errorf("Expecting no error but found err %v", err)
		}
	})
}

func setupCb(t *testing.T) (provider, sqlmock.Sqlmock, func() error) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	db := sqlx.NewDb(mockDB, "sqlmock")
	p := provider{
		driver:     "mssql",
		datasource: "connection string",
		db:         db,
		dialect:    mockStruct{},
		config: Config{
			DbName:     "NOCBO",
			Server:     "10.2.27.41",
			Password:   "its",
			UserID:     "its",
			Driver:     "mockMssql",
			CacheLimit: 200,
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled:                true,
					TimeoutInSecond:        3,
					MaxConcurrentRequests:  15,
					ErrorPercentThreshold:  25,
					RequestVolumeThreshold: 50,
					SleepWindowInSecond:    10,
				},
			},
		},
	}

	return p, mock, mockDB.Close
}

func TestExec(t *testing.T) {
	p, mock, closeDB := setupCb(t)
	defer closeDB()
	t.Run("Error running exec query", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer mockDB.Close()
		db := sqlx.NewDb(mockDB, "sqlmock")
		mock.ExpectExec("Insert into Ticket_Tracing values(10,GETDATE(),GETDATE())").WillReturnError(fmt.Errorf("some error"))

		p := provider{
			driver:     "mssql",
			datasource: "connection string",
			db:         db,
			config: Config{
				DbName:     "NOCBO",
				Server:     "10.2.27.41",
				Password:   "its",
				UserID:     "its",
				Driver:     "mockMssql",
				CacheLimit: 200,
				CircuitBreaker: CircuitBreaker{
					Config: &circuit.Config{
						Enabled: false,
					},
				},
			},
			dialect: mockStruct{},
		}

		err = p.Exec("Insert into Ticket_Tracing values(10,GETDATE(),GETDATE())")
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}
	})

	t.Run("Success exec query when circit breaker is enabled and no circuit breaker or db error is returned", func(t *testing.T) {
		mock.ExpectExec("Update Ticket set ID = 1").WillReturnResult(sqlmock.NewResult(1, 1))

		err := p.Exec("Update Ticket set ID = 1")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}
	})

	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := fmt.Errorf("Valid circuit breaker error : connection refused")

		mock.ExpectExec("Update Ticket set ID = 1").WillReturnError(errors.New("connection refused"))

		gotErr := p.Exec("Update Ticket set ID = 1")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}
	})
	t.Run("Invalid cb error", func(t *testing.T) {
		expectedErr := fmt.Errorf("duplicate_column")

		mock.ExpectExec("Update Ticket set ID = 1").WillReturnError(errors.New("duplicate_column"))

		gotErr := p.Exec("Update Ticket set ID = 1")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}
	})
	t.Run("Success exec query when circuit breaker is not enabled", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer mockDB.Close()
		db := sqlx.NewDb(mockDB, "sqlmock")
		mock.ExpectExec("Update Ticket set ID = 1").WillReturnResult(sqlmock.NewResult(1, 1))

		p := provider{
			driver:     "mssql",
			datasource: "connection string",
			db:         db,
			dialect:    mockStruct{},
			config: Config{
				DbName:     "NOCBO",
				Server:     "10.2.27.41",
				Password:   "its",
				UserID:     "its",
				Driver:     "mockMssql",
				CacheLimit: 200,
				CircuitBreaker: CircuitBreaker{
					Config: &circuit.Config{
						Enabled: false,
					},
				},
			},
		}

		err = p.Exec("Update Ticket set ID = 1")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}
	})
}

func TestSelect(t *testing.T) {
	p, mock, closeDB := setupCb(t)
	defer closeDB()
	t.Run("Select error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer mockDB.Close()
		db := sqlx.NewDb(mockDB, "sqlmock")
		mock.ExpectQuery("Select * from Ticket_Tracing where TicketID = 1").WillReturnError(fmt.Errorf("some error"))

		p := provider{
			driver:     "mssql",
			datasource: "connection string",
			db:         db,
			config: Config{
				CircuitBreaker: CircuitBreaker{
					Config: &circuit.Config{
						Enabled: false,
					},
				},
			},
			dialect: mockStruct{},
		}
		_, err = p.Select("Select * from Ticket_Tracing where TicketID = 1")
		if err == nil {
			t.Errorf("Expecting err but found nil")
		}
	})
	t.Run("Select success: when circuit breaker is not enabled", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer mockDB.Close()
		db := sqlx.NewDb(mockDB, "sqlmock")
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mock.ExpectQuery("Select Id, InTime, OutTime from Tracing").WillReturnRows(rows)

		p := provider{
			driver:     "mssql",
			datasource: "connection string",
			db:         db,
			config: Config{
				CircuitBreaker: CircuitBreaker{
					Config: &circuit.Config{
						Enabled: false,
					},
				},
			},
			dialect: mockStruct{},
		}
		_, err = p.Select("Select Id, InTime, OutTime from Tracing")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}
	})
	t.Run("Select success: circuit breaker enabled with no circuit breaker and db error.", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mock.ExpectQuery("Select Id, InTime, OutTime from Tracing").WillReturnRows(rows)

		_, err := p.Select("Select Id, InTime, OutTime from Tracing")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}
	})
	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("Valid circuit breaker error : connection refused")

		mock.ExpectQuery("Select Id, InTime, OutTime from Tracing").WillReturnError(errors.New("connection refused"))

		_, gotErr := p.Select("Select Id, InTime, OutTime from Tracing")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}
	})
	t.Run("Invalid cb error", func(t *testing.T) {
		expectedErr := errors.New("duplicate_column")

		mock.ExpectQuery("Select Id, InTime, OutTime from Tracing").WillReturnError(errors.New("duplicate_column"))

		_, gotErr := p.Select("Select Id, InTime, OutTime from Tracing")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}
	})
}

func TestConvertSqlRowstoMap(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer mockDB.Close()
		db := sqlx.NewDb(mockDB, "sqlmock")

		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mock.ExpectQuery("Select * from Tracing").WillReturnRows(rows)
		row, _ := db.Queryx("Select *s from Tracing")
		_, err = convertSQLRowsToMap(row)
		if err == nil {
			t.Errorf("expected error but found nil")
		}
	})
}

func TestCloseStatement(t *testing.T) {
	t.Run("success closing statement: when circuit breaker is enabled", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer mockDB.Close()
		db := sqlx.NewDb(mockDB, "sqlmock")

		getStatement = func(key string) (stmt *sqlx.Stmt) {
			st, _ := db.Preparex("Insert into Ticket values(?)")
			return st
		}

		mock.ExpectPrepare("Insert into").WillBeClosed()

		p := provider{
			driver:     "mssql",
			datasource: "connection string",
			db:         db,
			config: Config{
				DbName:     "NOCBO",
				Server:     "10.2.27.41",
				Password:   "its",
				UserID:     "its",
				Driver:     "mockMssql",
				CacheLimit: 200,
				CircuitBreaker: CircuitBreaker{
					Config: &circuit.Config{
						Enabled:                true,
						TimeoutInSecond:        3,
						MaxConcurrentRequests:  15,
						ErrorPercentThreshold:  25,
						RequestVolumeThreshold: 50,
						SleepWindowInSecond:    10,
					},
				},
			},
			dialect: mockStruct{},
		}
		err = p.CloseStatement("Insert into Ticket values(?)")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
	t.Run("success closing statement when cb is not enabled.", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer mockDB.Close()
		db := sqlx.NewDb(mockDB, "sqlmock")

		getStatement = func(key string) (stmt *sqlx.Stmt) {
			st, _ := db.Preparex("Insert into Ticket values(?)")
			return st
		}

		mock.ExpectPrepare("Insert into").WillBeClosed()

		p := provider{
			driver:     "mssql",
			datasource: "connection string",
			db:         db,
			config: Config{
				CircuitBreaker: CircuitBreaker{
					Config: &circuit.Config{
						Enabled: false,
					},
				},
			},
			dialect: mockStruct{},
		}

		err = p.CloseStatement("Insert into Ticket values(?)")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestSelectWithPrepare(t *testing.T) {
	initializeCache(Config{CacheLimit: 10})
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")
	p := provider{
		driver:     "mssql",
		datasource: "connection string",
		db:         db,
		config: Config{
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled: false,
				},
			},
		},
		dialect: mockStruct{},
	}

	cbP, mockDb, closeDB := setupCb(t)
	defer closeDB()

	t.Run("error", func(t *testing.T) {
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("Some Errors"))

		_, err = p.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Success: when circuit breaker is not enabled", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)

		_, err = p.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Success: when circuit breaker is enabled and no circuit breaker or db error is present.", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mockDb.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)
		_, err = cbP.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})

	t.Run("Error cached prepared statement", func(t *testing.T) {
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("Some Errors"))
		_, err = p.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})

	t.Run("Success cached prepared statement", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)
		_, err = p.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})

	t.Run("error creating prepared statement", func(t *testing.T) {
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("Some Error"))
		_, err = p.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("Failed to prepare a statement. Error: Valid circuit breaker error : connection refused")

		mockDb.ExpectPrepare("Select from Ticket where ID = ?").WillReturnError(errors.New("connection refused"))
		_, gotErr := cbP.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Invalid cb error", func(t *testing.T) {
		expectedErr := errors.New("Failed to prepare a statement. Error: duplicate_column")

		mockDb.ExpectPrepare("Select from Ticket where ID = ?").WillReturnError(errors.New("duplicate_column"))
		_, gotErr := cbP.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
}

func TestExecWithPrepare(t *testing.T) {
	initializeCache(Config{CacheLimit: 10})
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")
	p := provider{
		driver:     "mssql",
		datasource: "connection string",
		db:         db,
		dialect:    mockStruct{},
		config: Config{
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled: false,
				},
			},
		},
	}

	cbP, mockDb, closeDB := setupCb(t)
	defer closeDB()

	t.Run("Success: when circuit breaker is not enabled", func(t *testing.T) {
		mock.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnResult(sqlmock.NewResult(1, 1))

		if err = p.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2"); err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Success: when circuit breaker is enabled and no circit breaker or db error is present.", func(t *testing.T) {
		mockDb.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnResult(sqlmock.NewResult(1, 1))
		if err = cbP.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2"); err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Error", func(t *testing.T) {
		mock.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnError(errors.New("Some Error"))
		if err = p.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2"); err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})

	t.Run("Error cached statement", func(t *testing.T) {
		mock.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnError(errors.New("Some Error"))
		if err = p.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2"); err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})

	t.Run("Success cached statement", func(t *testing.T) {
		mock.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnResult(sqlmock.NewResult(1, 1))
		if err = p.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2"); err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})

	t.Run("Error creating prepared statement", func(t *testing.T) {
		mock.ExpectPrepare("INSERT INTO Ticket VALUES(?)").WillReturnError(errors.New("Some Error"))
		if err = p.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2"); err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("Valid circuit breaker error : connection refused")

		mockDb.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnError(errors.New("connection refused"))
		gotErr := cbP.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2")
		if gotErr == nil {
			t.Errorf("Expecting nil but found err %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})

	t.Run("Invalid cb error", func(t *testing.T) {
		expectedErr := errors.New("duplicate_column")

		mockDb.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnError(errors.New("duplicate_column"))
		gotErr := cbP.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2")
		if gotErr == nil {
			t.Errorf("Expecting nil but found err %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
}

func TestSelectAndProcess(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")
	p := provider{
		driver:     "mssql",
		datasource: "connection string",
		db:         db,
		dialect:    mockStruct{},
		config: Config{
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled: false,
				},
			},
		},
	}

	cbP, mockDb, closeDB := setupCb(t)
	defer closeDB()

	t.Run("select error", func(t *testing.T) {
		isError = false

		mock.ExpectQuery("from Ticket_Tracing where TicketID").WillReturnError(fmt.Errorf("some error"))
		p.SelectAndProcess("Select * from Ticket_Tracing where TicketID = 1", process)
		if !isError {
			t.Errorf("Expecting err but found nil")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("select success: when circuit breaker is not enabled", func(t *testing.T) {
		isError = false
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mock.ExpectQuery("Select Id, InTime, OutTime from Tracing").WillReturnRows(rows)

		p.SelectAndProcess("Select Id, InTime, OutTime from Tracing", process)
		if isError {
			t.Errorf("Expecting nil but found err")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("select success: when circuit breaker is enabled", func(t *testing.T) {
		isError = false
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mockDb.ExpectQuery("Select Id, InTime, OutTime from Tracing").WillReturnRows(rows)

		cbP.SelectAndProcess("Select Id, InTime, OutTime from Tracing", process)
		if isError {
			t.Errorf("Expecting nil but found err")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("Valid circuit breaker error : connection refused")
		var errorReceivedInCallback Row
		mockProcessRowCallback := func(row Row) {
			errorReceivedInCallback = Row{
				Error: expectedErr,
			}
		}

		isError = true
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mockDb.ExpectQuery("Select Id, InTime, OutTime from Tracing").WillReturnRows(rows)

		cbP.SelectAndProcess("Select Id, InTime, OutTime from Tracing", mockProcessRowCallback)
		if !isError {
			t.Errorf("Expecting err but found nil")
		}

		if !reflect.DeepEqual(errorReceivedInCallback.Error, expectedErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, errorReceivedInCallback.Error)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Invalid cb error", func(t *testing.T) {
		expectedErr := errors.New("too_many_rows")
		var errorReceivedInCallback Row
		mockProcessRowCallback := func(row Row) {
			errorReceivedInCallback = Row{
				Error: expectedErr,
			}
		}

		isError = true
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mockDb.ExpectQuery("Select Id, InTime, OutTime from Tracing").WillReturnRows(rows)

		cbP.SelectAndProcess("Select Id, InTime, OutTime from Tracing", mockProcessRowCallback)
		if !isError {
			t.Errorf("Expecting err but found nil")
		}

		if !reflect.DeepEqual(errorReceivedInCallback.Error, expectedErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, errorReceivedInCallback.Error)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSelectWithPrepareAndProcess(t *testing.T) {
	initializeCache(Config{CacheLimit: 10})
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")
	p := provider{
		driver:     "mssql",
		datasource: "connection string",
		db:         db,
		dialect:    mockStruct{},
		config: Config{
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled: false,
				},
			},
		},
	}

	cbP, mockDb, closeDB := setupCb(t)
	defer closeDB()

	t.Run("success: when circuit breaker is not enabled", func(t *testing.T) {
		isError = false
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)

		p.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", process, "2")
		if isError {
			t.Errorf("Expecting nil but found err")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("success: when circuit breaker is enabled", func(t *testing.T) {
		isError = false
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mockDb.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)

		cbP.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", process, "2")
		if isError {
			t.Errorf("Expecting nil but found err")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("Valid circuit breaker error : connection refused")
		var errorReceivedInCallback Row
		mockProcessRowCallback := func(row Row) {
			errorReceivedInCallback = Row{
				Error: expectedErr,
			}
		}

		isError = true
		mockDb.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("connection refused"))

		cbP.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", mockProcessRowCallback, "2")
		if !isError {
			t.Errorf("Expecting err but found nil")
		}

		if !reflect.DeepEqual(errorReceivedInCallback.Error, expectedErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, errorReceivedInCallback.Error)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Invalid cb error", func(t *testing.T) {
		expectedErr := errors.New("too_many_rows")
		var errorReceivedInCallback Row
		mockProcessRowCallback := func(row Row) {
			errorReceivedInCallback = Row{
				Error: expectedErr,
			}
		}

		isError = true
		errorReceivedInCallback.Error = nil
		mockDb.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("duplicate_column"))

		cbP.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", mockProcessRowCallback, "2")
		if !isError {
			t.Errorf("Expecting err but found nil")
		}

		if !reflect.DeepEqual(errorReceivedInCallback.Error, expectedErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, errorReceivedInCallback.Error)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Error cached prepared statement", func(t *testing.T) {
		isError = true
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("Some Errors"))

		p.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", process, "2")
		if !isError {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})

	t.Run("Success cached prepared statement", func(t *testing.T) {
		isError = false
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)
		p.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", process, "2")
		if isError {
			t.Errorf("Expecting nil but found err ")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})

	t.Run("error creating prepared statement", func(t *testing.T) {
		isError = false
		mock.ExpectPrepare("Select from Ticket where ID = ?").WillReturnError(errors.New("Some Error"))
		p.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", process, "2")
		if !isError {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
}

func TestSelectObjectsWithPrepare(t *testing.T) {
	initializeCache(Config{CacheLimit: 10})
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")
	p := provider{
		driver:     "postgres",
		datasource: "connection string",
		db:         db,
		dialect:    mockStruct{},
		config: Config{
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled: false,
				},
			},
		},
	}

	cbP, mockDb, closeDB := setupCb(t)
	defer closeDB()

	t.Run("error", func(t *testing.T) {
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("Some Errors"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}

		err = p.SelectObjectsWithPrepare("tran-1", &entities, "Select from Ticket where ID = ?", "2")
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}
		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("success: when circuit breaker is not enabled", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = p.SelectObjectsWithPrepare("tran-2", &entities, "Select from Ticket where ID = ?", "2")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if len(entities) != 2 {
			t.Errorf("Expecting 2 records returned but found %d", len(entities))
		}

		for _, entity := range entities {
			if entity.ID == 1 {
				if entity.Intime != "10 Aug" || entity.Outtime != "11 Aug" {
					t.Error("Expected returned record values do not match for id 1")
				}
			} else if entity.ID == 2 {
				if entity.Intime != "12 Aug" || entity.Outtime != "13 Aug" {
					t.Error("Expected returned record values do not match for id 2")
				}
			} else {
				t.Error("Expected returned record IDs do not match")
			}
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("success: when circuit breaker is enabled and no circuit breaker or db error is returned.", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mockDb.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = cbP.SelectObjectsWithPrepare("tran-2", &entities, "Select from Ticket where ID = ?", "2")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if len(entities) != 2 {
			t.Errorf("Expecting 2 records returned but found %d", len(entities))
		}

		for _, entity := range entities {
			if entity.ID == 1 {
				if entity.Intime != "10 Aug" || entity.Outtime != "11 Aug" {
					t.Error("Expected returned record values do not match for id 1")
				}
			} else if entity.ID == 2 {
				if entity.Intime != "12 Aug" || entity.Outtime != "13 Aug" {
					t.Error("Expected returned record values do not match for id 2")
				}
			} else {
				t.Error("Expected returned record IDs do not match")
			}
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})

	t.Run("Error cached prepared statement", func(t *testing.T) {
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("Some Errors"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = p.SelectObjectsWithPrepare("tran-3", &entities, "Select from Ticket where ID = ?", "2")
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})

	t.Run("Success cached prepared statement", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = p.SelectObjectsWithPrepare("tran-4", &entities, "Select from Ticket where ID = ?", "2")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if len(entities) != 2 {
			t.Errorf("Expecting 2 records returned but found %d", len(entities))
		}

		for _, entity := range entities {
			if entity.ID == 1 {
				if entity.Intime != "10 Aug" || entity.Outtime != "11 Aug" {
					t.Error("Expected returned record values do not match for id 1")
				}
			} else if entity.ID == 2 {
				if entity.Intime != "12 Aug" || entity.Outtime != "13 Aug" {
					t.Error("Expected returned record values do not match for id 2")
				}
			} else {
				t.Error("Expected returned record IDs do not match")
			}
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})

	t.Run("error creating prepared statement", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer mockDB.Close()
		db := sqlx.NewDb(mockDB, "sqlmock")

		old := getStatement
		getStatement = func(key string) (stmt *sqlx.Stmt) {
			return nil
		}

		defer func() {
			getStatement = old
		}()

		mock.ExpectPrepare("Select from Ticket where ID = ?").WillReturnError(errors.New("Some Error"))
		p := provider{
			driver:     "postgres",
			datasource: "connection string",
			db:         db,
			dialect:    mockStruct{},
			config: Config{
				CircuitBreaker: CircuitBreaker{
					Config: &circuit.Config{
						Enabled: false,
					},
				},
			},
		}
		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = p.SelectObjectsWithPrepare("tran-5", &entities, "Select from Ticket where ID = ?", "2")
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("Valid circuit breaker error : connection refused")

		mockDb.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("connection refused"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		gotErr := cbP.SelectObjectsWithPrepare("tran-2", &entities, "Select from Ticket where ID = ?", "2")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Invalid cb error", func(t *testing.T) {
		expectedErr := errors.New("duplicate_column")

		mockDb.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("duplicate_column"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		gotErr := cbP.SelectObjectsWithPrepare("tran-2", &entities, "Select from Ticket where ID = ?", "2")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
}

func TestSelectObjects(t *testing.T) {
	initializeCache(Config{CacheLimit: 10})
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")
	p := provider{
		driver:     "postgres",
		datasource: "connection string",
		db:         db,
		dialect:    mockStruct{},
		config: Config{
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled: false,
				},
			},
		},
	}

	cbP, mockDb, closeDB := setupCb(t)
	defer closeDB()

	t.Run("error", func(t *testing.T) {
		mock.ExpectQuery("Select from Ticket").WillReturnError(errors.New("Some Errors"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}

		err = p.SelectObjects("tran-1", &entities, "Select from Ticket")
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}
		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("success: when circuit breaker is not enabled", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mock.ExpectQuery("Select from Ticket").WillReturnRows(rows)
		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = p.SelectObjects("tran-2", &entities, "Select from Ticket")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if len(entities) != 2 {
			t.Errorf("Expecting 2 records returned but found %d", len(entities))
		}

		for _, entity := range entities {
			if entity.ID == 1 {
				if entity.Intime != "10 Aug" || entity.Outtime != "11 Aug" {
					t.Error("Expected returned record values do not match for id 1")
				}
			} else if entity.ID == 2 {
				if entity.Intime != "12 Aug" || entity.Outtime != "13 Aug" {
					t.Error("Expected returned record values do not match for id 2")
				}
			} else {
				t.Error("Expected returned record IDs do not match")
			}
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("success: circuit breaker enabled with no circuit breaker or db error.", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mockDb.ExpectQuery("Select from Ticket").WillReturnRows(rows)
		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = cbP.SelectObjects("tran-2", &entities, "Select from Ticket")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if len(entities) != 2 {
			t.Errorf("Expecting 2 records returned but found %d", len(entities))
		}

		for _, entity := range entities {
			if entity.ID == 1 {
				if entity.Intime != "10 Aug" || entity.Outtime != "11 Aug" {
					t.Error("Expected returned record values do not match for id 1")
				}
			} else if entity.ID == 2 {
				if entity.Intime != "12 Aug" || entity.Outtime != "13 Aug" {
					t.Error("Expected returned record values do not match for id 2")
				}
			} else {
				t.Error("Expected returned record IDs do not match")
			}
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("Valid circuit breaker error : connection refused")

		mockDb.ExpectQuery("Select from Ticket").WillReturnError(errors.New("connection refused"))
		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		gotErr := cbP.SelectObjects("tran-2", &entities, "Select from Ticket")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("Invalid cb error", func(t *testing.T) {
		expectedErr := errors.New("duplicate_column")

		mockDb.ExpectQuery("Select from Ticket").WillReturnError(errors.New("duplicate_column"))
		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		gotErr := cbP.SelectObjects("tran-2", &entities, "Select from Ticket")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
	t.Run("success with no records", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"})
		mock.ExpectQuery("Select from Ticket").WillReturnRows(rows)

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = p.SelectObjects("tran-2", &entities, "Select from Ticket")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		flush()
	})
}

func TestPrepareStatement(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")
	p := provider{
		driver:     "postgres",
		datasource: "connection string",
		db:         db,
		dialect:    mockStruct{},
		config: Config{
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled: false,
				},
			},
		},
	}

	cbP := provider{
		driver:     "mssql",
		datasource: "connection string",
		db:         db,
		dialect:    mockStruct{},
		config: Config{
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled:                true,
					TimeoutInSecond:        3,
					MaxConcurrentRequests:  15,
					ErrorPercentThreshold:  25,
					RequestVolumeThreshold: 50,
					SleepWindowInSecond:    10,
				},
			},
		},
	}

	addStatementMockCalled := false
	inputQuery := "select from test"
	injectedError := errors.New("injected-error")
	callbackError := errors.New("Failed to prepare a statement. Error: injected-error")

	defer func(f func(string) (stmt *sqlx.Stmt)) { getStatement = f }(getStatement)
	defer func(f func(string, string, *sqlx.Stmt)) { addStatement = f }(addStatement)

	addStatement = func(string, string, *sqlx.Stmt) {
		addStatementMockCalled = true
	}

	type args struct {
		scenario             string
		expectedResult       error
		addStatementExpected bool
		setupMocks           func()
	}

	tests := []args{
		{
			scenario: "Statement exists",
			setupMocks: func() {
				getStatement = func(string) *sqlx.Stmt {
					return &sqlx.Stmt{}
				}
			},
			addStatementExpected: true,
		},
		{
			scenario: "Statement doesn't exist, prepare passes",
			setupMocks: func() {
				getStatement = func(string) *sqlx.Stmt {
					return nil
				}
				mock.ExpectPrepare(inputQuery)
			},
			addStatementExpected: true,
		},
		{
			scenario: "Statement doesn't exist, prepare fails",
			setupMocks: func() {
				getStatement = func(string) *sqlx.Stmt {
					return nil
				}
				mock.ExpectPrepare(inputQuery).WillReturnError(injectedError)
			},
			expectedResult: callbackError,
		},
		{
			scenario: "Statement exists: when cicuit breaker is enabled",
			setupMocks: func() {
				getStatement = func(string) *sqlx.Stmt {
					return &sqlx.Stmt{}
				}
			},
			addStatementExpected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			addStatementMockCalled = false
			test.setupMocks()
			var err error

			if test.scenario == "Statement exists: when cicuit breaker is enabled" {
				err = cbP.PrepareStatement("transactionID-1", inputQuery)
			} else {
				err = p.PrepareStatement("transactionID-1", inputQuery)
			}

			if !reflect.DeepEqual(err, test.expectedResult) {
				t.Errorf("Expected error: %v but got error: %v", test.expectedResult, err)
			}

			if addStatementMockCalled != test.addStatementExpected {
				t.Fatalf("addStatement mock call expectation was set to %t but got %t", test.addStatementExpected, addStatementMockCalled)
			}
			if e := mock.ExpectationsWereMet(); e != nil {
				t.Fatalf("some expectations were not met, %v", e)
			}
		})
	}
}

func TestSelectObjectAndProcess(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")
	p := provider{
		driver:     "postgres",
		datasource: "connection string",
		db:         db,
		dialect:    mockStruct{},
		config: Config{
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled: true,
				},
			},
		},
	}
	cbP := provider{
		driver:     "mssql",
		datasource: "connection string",
		db:         db,
		dialect:    mockStruct{},
		config: Config{
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled:                true,
					TimeoutInSecond:        3,
					MaxConcurrentRequests:  15,
					ErrorPercentThreshold:  25,
					RequestVolumeThreshold: 50,
					SleepWindowInSecond:    10,
				},
			},
		},
	}

	var errorReceivedInCallback error
	callbackInvokeCount := 0
	sendErrorFromCallback := false
	inputQuery := "select from test"
	injectedError := errors.New("injected-error")
	validCbError := errors.New("Valid circuit breaker error : connection refused")
	invalidCbError := errors.New("too_many_rows")

	mockProcessObjectCallback := func(object interface{}, err error) error {
		callbackInvokeCount++
		if errorReceivedInCallback == nil {
			errorReceivedInCallback = err
		}
		if sendErrorFromCallback {
			return injectedError
		}
		return nil
	}

	type entity struct {
		ID      int    `db:"Id"`
		Intime  string `db:"InTime"`
		Outtime string `db:"OutTime"`
	}

	type args struct {
		scenario                        string
		expectedCallbackInvocationCount int
		expectedCallbackError           error
		setupMocks                      func()
	}

	tests := []args{

		{
			scenario: "Success: when circuit breaker is not enabled",
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				mock.ExpectQuery(inputQuery).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 2,
		},
		{
			scenario: "Success: when circuit breaker is enabled and no circuit breaker or db error is returned.",
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				mock.ExpectQuery(inputQuery).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 2,
		},
		{
			scenario: "Query failed",
			setupMocks: func() {
				mock.ExpectQuery(inputQuery).WillReturnError(injectedError)
			},
			expectedCallbackInvocationCount: 1,
			expectedCallbackError:           injectedError,
		},
		{
			scenario: "Row scan failed",
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				rows = rows.RowError(1, injectedError)
				mock.ExpectQuery(inputQuery).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 2,
			expectedCallbackError:           injectedError,
		},

		{
			scenario: "Callback error",
			setupMocks: func() {
				sendErrorFromCallback = true
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				mock.ExpectQuery(inputQuery).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 1,
		},
		{
			scenario: "Valid cb error",
			setupMocks: func() {
				mock.ExpectQuery(inputQuery).WillReturnError(errors.New("connection refused"))
			},
			expectedCallbackInvocationCount: 1,
			expectedCallbackError:           validCbError,
		},
		{
			scenario: "Invalid cb error",
			setupMocks: func() {
				mock.ExpectQuery(inputQuery).WillReturnError(invalidCbError)
			},
			expectedCallbackInvocationCount: 1,
			expectedCallbackError:           invalidCbError,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			errorReceivedInCallback = nil
			callbackInvokeCount = 0
			sendErrorFromCallback = false

			test.setupMocks()

			if test.scenario == "Success: when circuit breaker is enabled" || test.scenario == "Valid cb error" || test.scenario == "Invalid cb error" {
				cbP.SelectObjectAndProcess("transaction-1", &entity{}, mockProcessObjectCallback, inputQuery)
			} else {
				p.SelectObjectAndProcess("transaction-1", &entity{}, mockProcessObjectCallback, inputQuery)
			}

			if !reflect.DeepEqual(errorReceivedInCallback, test.expectedCallbackError) {
				t.Errorf("Expected error: %v but got error: %v", test.expectedCallbackError, errorReceivedInCallback)
			}

			if callbackInvokeCount != test.expectedCallbackInvocationCount {
				t.Fatalf("expected callback to be called %d times but got %d times", test.expectedCallbackInvocationCount, callbackInvokeCount)
			}
			if e := mock.ExpectationsWereMet(); e != nil {
				t.Fatalf("some expectations were not met: %v", e)
			}
		})
	}
}

func TestSelectObjectWithPrepareAndProcess(t *testing.T) {
	initializeCache(Config{CacheLimit: 10})
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")
	p := provider{
		driver:     "postgres",
		datasource: "connection string",
		db:         db,
		dialect:    mockStruct{},
		config: Config{
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled: false,
				},
			},
		},
	}
	cbP := provider{
		driver:     "mssql",
		datasource: "connection string",
		db:         db,
		dialect:    mockStruct{},
		config: Config{
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled:                true,
					TimeoutInSecond:        3,
					MaxConcurrentRequests:  15,
					ErrorPercentThreshold:  25,
					RequestVolumeThreshold: 50,
					SleepWindowInSecond:    10,
				},
			},
		},
	}

	var errorReceivedInCallback error
	callbackInvokeCount := 0
	sendErrorFromCallback := false
	inputQuery := "select from test"
	inputBindVariable := 1
	injectedError := errors.New("injected-error")
	callbackError := errors.New("Failed to prepare a statement. Error: injected-error")
	validCbError := errors.New("Valid circuit breaker error : connection refused")
	invalidCbError := errors.New("too_many_rows")

	mockProcessObjectCallback := func(object interface{}, err error) error {
		callbackInvokeCount++
		if errorReceivedInCallback == nil {
			errorReceivedInCallback = err
		}
		if sendErrorFromCallback {
			return injectedError
		}
		return nil
	}

	type entity struct {
		ID      int    `db:"Id"`
		Intime  string `db:"InTime"`
		Outtime string `db:"OutTime"`
	}

	type args struct {
		scenario                        string
		expectedCallbackInvocationCount int
		expectedCallbackError           error
		setupMocks                      func()
	}

	tests := []args{

		{
			scenario: "Success: when circuit breaker is not enabled",
			setupMocks: func() {
				mock.ExpectPrepare(inputQuery)
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 2,
		},
		{
			scenario: "Success: when circuit breaker is enabled and no circuit breaker or db error is returned.",
			setupMocks: func() {
				mock.ExpectPrepare(inputQuery)
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 2,
		},
		{
			scenario: "Valid cb error",
			setupMocks: func() {
				mock.ExpectPrepare(inputQuery)
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnError(errors.New("connection refused"))
			},
			expectedCallbackInvocationCount: 1,
			expectedCallbackError:           validCbError,
		},
		{
			scenario: "Invalid cb error",
			setupMocks: func() {
				mock.ExpectPrepare(inputQuery)
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				rows = rows.RowError(1, invalidCbError)
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 2,
			expectedCallbackError:           invalidCbError,
		},
		{
			scenario: "Query prepare failed",
			setupMocks: func() {
				mock.ExpectPrepare(inputQuery).WillReturnError(injectedError)
			},
			expectedCallbackError:           callbackError,
			expectedCallbackInvocationCount: 1,
		},

		{
			scenario: "Query failed",
			setupMocks: func() {
				mock.ExpectPrepare(inputQuery)
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnError(injectedError)
			},
			expectedCallbackInvocationCount: 1,
			expectedCallbackError:           injectedError,
		},

		{
			scenario: "Row scan failed",
			setupMocks: func() {
				mock.ExpectPrepare(inputQuery)
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				rows = rows.RowError(1, injectedError)
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 2,
			expectedCallbackError:           injectedError,
		},

		{
			scenario: "Callback error",
			setupMocks: func() {
				mock.ExpectPrepare(inputQuery)
				sendErrorFromCallback = true
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			errorReceivedInCallback = nil
			callbackInvokeCount = 0
			sendErrorFromCallback = false

			test.setupMocks()

			if test.scenario == "Success: when circuit breaker is enabled and no circuit breaker or db error is returned." || test.scenario == "Valid cb error" || test.scenario == "Invalid cb error" {
				cbP.SelectObjectWithPrepareAndProcess("transaction-1", &entity{}, mockProcessObjectCallback, inputQuery, inputBindVariable)
			} else {
				p.SelectObjectWithPrepareAndProcess("transaction-1", &entity{}, mockProcessObjectCallback, inputQuery, inputBindVariable)
			}

			if !reflect.DeepEqual(errorReceivedInCallback, test.expectedCallbackError) {
				t.Errorf("Expected error: %v but got error: %v", test.expectedCallbackError, errorReceivedInCallback)
			}

			// if errorReceivedInCallback != test.expectedCallbackError {
			// 	t.Fatalf("expected %v error in callback but got %v", test.expectedCallbackError, errorReceivedInCallback)
			// }

			if callbackInvokeCount != test.expectedCallbackInvocationCount {
				t.Fatalf("expected callback to be called %d times but got %d times", test.expectedCallbackInvocationCount, callbackInvokeCount)
			}
			if e := mock.ExpectationsWereMet(); e != nil {
				t.Fatalf("some expectations were not met: %v", e)
			}
			flush()
		})
	}
}
