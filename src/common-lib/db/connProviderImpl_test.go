package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

var postgresqlErrors = []string{
	"connection_exception", "connection_does_not_exist",
	"connection_failure", "bad connection", "connection refused",
	"sqlclient_unable_to_establish_sqlconnection",
	"sqlserver_rejected_establishment_of_sqlconnection", "broken pipe", "sql: database is closed",
}

func TestConnProviderExec(t *testing.T) {
	t.Run("Error running exec query CB Enabled", func(t *testing.T) {
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

		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		err = connProvider.Exec("Insert into Ticket_Tracing values(10,GETDATE(),GETDATE())")
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}
	})
	t.Run("success exec query CB Enabled", func(t *testing.T) {
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

		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		err = connProvider.Exec("Update Ticket set ID = 1")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}
	})

	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := fmt.Errorf("Valid circuit breaker error : connection refused")
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		mock.ExpectExec("Update Ticket set ID = 1").WillReturnError(errors.New("connection refused"))

		gotErr := connProvider.Exec("Update Ticket set ID = 1")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}
	})

	t.Run("Invalid cb error", func(t *testing.T) {
		expectedErr := fmt.Errorf("duplicate_column")
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		mock.ExpectExec("Update Ticket set ID = 1").WillReturnError(errors.New("duplicate_column"))

		gotErr := p.Exec("Update Ticket set ID = 1")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}
	})
}

func TestConnProviderSelect(t *testing.T) {
	t.Run("select error", func(t *testing.T) {
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

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		_, err = connProvider.Select("Select * from Ticket_Tracing where TicketID = 1")
		if err == nil {
			t.Errorf("Expecting err but found nil")
		}
	})

	t.Run("select success", func(t *testing.T) {
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

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		_, err = connProvider.Select("Select Id, InTime, OutTime from Tracing")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}
	})

	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("Valid circuit breaker error : connection refused")
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		mock.ExpectQuery("Select Id, InTime, OutTime from Tracing").WillReturnError(errors.New("connection refused"))

		_, gotErr := connProvider.Select("Select Id, InTime, OutTime from Tracing")
		if gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}
	})
	t.Run("Invalid cb error", func(t *testing.T) {
		expectedErr := errors.New("duplicate_column")
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
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

func TestConnProviderCloseStatement(t *testing.T) {
	t.Run("success closing statement", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer mockDB.Close()
		db := sqlx.NewDb(mockDB, "sqlmock")

		mock.ExpectPrepare("Insert into").WillBeClosed()

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

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		connProvider.PrepareStatement("", "Insert into")

		err = connProvider.CloseStatement("Insert into")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestConnProviderSelectWithPrepare(t *testing.T) {
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

	ctx := context.Background()

	t.Run("error", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("Some Errors"))

		_, err = connProvider.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("success", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)
		_, err = connProvider.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Error cached prepared statement", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("Some Errors"))
		_, err = connProvider.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Success cached prepared statement", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)
		_, err = connProvider.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error creating prepared statement", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		mock.ExpectPrepare("Select from Ticket where ID = ?").WillReturnError(errors.New("Some Error"))
		_, err = connProvider.SelectWithPrepare("Select from Ticket where ID = ?", "2")
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("connection refused")
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		mock.ExpectPrepare("Select from Ticket where ID = ?").WillReturnError(errors.New("connection refused"))
		_, gotErr := connProvider.SelectWithPrepare("Select from Ticket where ID = ?", "2")
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
		expectedErr := errors.New("duplicate_column")
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		mock.ExpectPrepare("Select from Ticket where ID = ?").WillReturnError(errors.New("duplicate_column"))
		_, gotErr := connProvider.SelectWithPrepare("Select from Ticket where ID = ?", "2")
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

func TestConnProviderExecWithPrepare(t *testing.T) {
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

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		mock.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnResult(sqlmock.NewResult(1, 1))

		if err = connProvider.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2"); err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Error", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		mock.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnError(errors.New("Some Error"))
		if err = connProvider.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2"); err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Error cached statement", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		mock.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnError(errors.New("Some Error"))
		if err = connProvider.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2"); err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Success cached statement", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		mock.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnResult(sqlmock.NewResult(1, 1))
		if err = connProvider.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2"); err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Error creating prepared statement", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		mock.ExpectPrepare("INSERT INTO Ticket VALUES(?)").WillReturnError(errors.New("Some Error"))
		if err = connProvider.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2"); err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("Valid circuit breaker error : connection refused")
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		mock.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnError(errors.New("connection refused"))
		gotErr := connProvider.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2")
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
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		mock.ExpectPrepare("INSERT INTO Ticket VALUES(?)").ExpectExec().WithArgs("2").WillReturnError(errors.New("duplicate_column"))
		gotErr := connProvider.ExecWithPrepare("INSERT INTO Ticket VALUES(?)", "2")
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

func TestConnProviderSelectAndProcess(t *testing.T) {
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

	ctx := context.Background()

	connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
	defer connProvider.Close("")

	t.Run("select error", func(t *testing.T) {
		isError = false

		mock.ExpectQuery("from Ticket_Tracing where TicketID").WillReturnError(fmt.Errorf("some error"))
		connProvider.SelectAndProcess("Select * from Ticket_Tracing where TicketID = 1", process)
		if !isError {
			t.Errorf("Expecting err but found nil")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("select success", func(t *testing.T) {
		isError = false
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mock.ExpectQuery("Select Id, InTime, OutTime from Tracing").WillReturnRows(rows)

		connProvider.SelectAndProcess("Select Id, InTime, OutTime from Tracing", process)
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
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		isError = true
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mock.ExpectQuery("Select Id, InTime, OutTime from Tracing").WillReturnRows(rows)

		connProvider.SelectAndProcess("Select Id, InTime, OutTime from Tracing", mockProcessRowCallback)
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
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		isError = true
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mock.ExpectQuery("Select Id, InTime, OutTime from Tracing").WillReturnRows(rows)

		connProvider.SelectAndProcess("Select Id, InTime, OutTime from Tracing", mockProcessRowCallback)
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

func TestConnProviderSelectWithPrepareAndProcess(t *testing.T) {
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

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		isError = false
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)

		connProvider.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", process, "2")
		if isError {
			t.Errorf("Expecting nil but found err")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Error cached prepared statement", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		isError = true
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("Some Errors"))

		connProvider.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", process, "2")
		if !isError {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Success cached prepared statement", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		isError = false
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)
		connProvider.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", process, "2")
		if isError {
			t.Errorf("Expecting nil but found err ")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error creating prepared statement", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		isError = false
		mock.ExpectPrepare("Select from Ticket where ID = ?").WillReturnError(errors.New("Some Error"))
		connProvider.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", process, "2")
		if !isError {
			t.Errorf("Expecting error but found nil")
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
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		isError = true
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("connection refused"))

		connProvider.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", mockProcessRowCallback, "2")
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
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("duplicate_column"))

		connProvider.SelectWithPrepareAndProcess("Select from Ticket where ID = ?", mockProcessRowCallback, "2")
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
}

func TestConnProviderSelectObjectsWithPrepare(t *testing.T) {
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

	ctx := context.Background()

	t.Run("error", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("Some Errors"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}

		err = connProvider.SelectObjectsWithPrepare("tran-1", &entities, "Select from Ticket where ID = ?", "2")
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
	t.Run("success", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = connProvider.SelectObjectsWithPrepare("tran-2", &entities, "Select from Ticket where ID = ?", "2")
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
	})

	t.Run("Error cached prepared statement", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("Some Errors"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = connProvider.SelectObjectsWithPrepare("tran-3", &entities, "Select from Ticket where ID = ?", "2")
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

	t.Run("Success cached prepared statement", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnRows(rows)

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = connProvider.SelectObjectsWithPrepare("tran-4", &entities, "Select from Ticket where ID = ?", "2")
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

		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = connProvider.SelectObjectsWithPrepare("tran-5", &entities, "Select from Ticket where ID = ?", "2")
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
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("connection refused"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		gotErr := connProvider.SelectObjectsWithPrepare("tran-2", &entities, "Select from Ticket where ID = ?", "2")
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
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		mock.ExpectPrepare("Select from Ticket where ID = ?").ExpectQuery().WithArgs("2").WillReturnError(errors.New("duplicate_column"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		gotErr := connProvider.SelectObjectsWithPrepare("tran-2", &entities, "Select from Ticket where ID = ?", "2")
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

func TestConnProviderSelectObjects(t *testing.T) {
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

	ctx := context.Background()

	t.Run("error", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		mock.ExpectQuery("Select from Ticket").WillReturnError(errors.New("Some Errors"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}

		err = connProvider.SelectObjects("tran-1", &entities, "Select from Ticket")
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
	t.Run("success", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mock.ExpectQuery("Select from Ticket").WillReturnRows(rows)
		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = connProvider.SelectObjects("tran-2", &entities, "Select from Ticket")
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
	})
	t.Run("success with no records", func(t *testing.T) {
		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")

		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"})
		mock.ExpectQuery("Select from Ticket").WillReturnRows(rows)

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		err = connProvider.SelectObjects("tran-2", &entities, "Select from Ticket")
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("connection refused")
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		mock.ExpectQuery("Select from Ticket").WillReturnError(errors.New("connection refused"))
		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		gotErr := connProvider.SelectObjects("tran-2", &entities, "Select from Ticket")
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
		p, mock, _ := setupCb(t)
		ctx := context.Background()

		connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
		defer connProvider.Close("")
		mock.ExpectQuery("Select from Ticket").WillReturnError(errors.New("duplicate_column"))
		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		gotErr := connProvider.SelectObjects("tran-2", &entities, "Select from Ticket")
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

func TestConnProviderPrepareStatement(t *testing.T) {
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

	ctx := context.Background()

	inputQuery := "select from test"
	injectedError := errors.New("injected-error")

	defer func(f func(string) (stmt *sqlx.Stmt)) { getStatement = f }(getStatement)

	type args struct {
		scenario       string
		expectedResult error
		setupMocks     func()
	}

	tests := []args{
		{
			scenario: "Success",
			setupMocks: func() {
				getStatement = func(string) *sqlx.Stmt {
					return nil
				}
				mock.ExpectPrepare(inputQuery)
			},
		},
		{
			scenario: "Prepare errors",
			setupMocks: func() {
				getStatement = func(string) *sqlx.Stmt {
					return nil
				}
				mock.ExpectPrepare(inputQuery).WillReturnError(injectedError)
			},
			expectedResult: injectedError,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
			defer connProvider.Close("")

			test.setupMocks()

			err := connProvider.PrepareStatement("transactionID-1", inputQuery)

			if err != test.expectedResult {
				t.Fatalf("expected result %+v but got %+v", test.expectedResult, err)
			}
			if e := mock.ExpectationsWereMet(); e != nil {
				t.Fatalf("some expectations were not met, %v", e)
			}
		})
	}
}

func TestConnProviderSelectObjectAndProcess(t *testing.T) {
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

	ctx := context.Background()

	connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
	defer connProvider.Close("")

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
			scenario: "Success",
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

			connProvider.SelectObjectAndProcess("transaction-1", &entity{}, mockProcessObjectCallback, inputQuery)

			if !reflect.DeepEqual(errorReceivedInCallback, test.expectedCallbackError) {
				t.Fatalf("expected %v error in callback but got %v", test.expectedCallbackError, errorReceivedInCallback)
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

func TestConnProviderSelectObjectWithPrepareAndProcess(t *testing.T) {
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

	ctx := context.Background()

	var errorReceivedInCallback error
	callbackInvokeCount := 0
	sendErrorFromCallback := false
	inputQuery := "select from test"
	inputBindVariable := 1
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
			scenario: "Success",
			setupMocks: func() {
				mock.ExpectPrepare(inputQuery)
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 2,
		},
		{
			scenario: "Query prepare failed",
			setupMocks: func() {
				mock.ExpectPrepare(inputQuery).WillReturnError(injectedError)
			},
			expectedCallbackError:           injectedError,
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
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			connProvider, _ := p.GetSingleConnectionProvider(ctx, "")
			defer connProvider.Close("")

			errorReceivedInCallback = nil
			callbackInvokeCount = 0
			sendErrorFromCallback = false

			test.setupMocks()

			connProvider.SelectObjectWithPrepareAndProcess("transaction-1", &entity{}, mockProcessObjectCallback, inputQuery, inputBindVariable)

			if !reflect.DeepEqual(errorReceivedInCallback, test.expectedCallbackError) {
				t.Fatalf("expected %v error in callback but got %v", test.expectedCallbackError, errorReceivedInCallback)
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
