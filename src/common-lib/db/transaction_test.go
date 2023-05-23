package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

const (
	inserQuery         = "Insert into Ticket_Tracing values(10,GETDATE(),GETDATE())"
	updateQuery        = "Update Ticket set ID = 1"
	selectQuery        = "Select from Ticket"
	insertPrepareQuery = "INSERT INTO Ticket VALUES(?)"
	selectPrepareQuery = "Select from Ticket where ID = ?"
)

func setup(t *testing.T) (provider, sqlmock.Sqlmock, func() error) {
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
			CircuitBreaker: CircuitBreaker{
				Config: &circuit.Config{
					Enabled: false,
				}},
		},
	}

	return p, mock, mockDB.Close
}

func TestBeginTransaction(t *testing.T) {
	p, mock, closeDB := setupCb(t)
	defer closeDB()

	t.Run("Success starting transaction", func(t *testing.T) {

		p, mock, closeDB := setup(t)
		defer closeDB()

		mock.ExpectBegin()
		ctx := context.Background()
		dbtx, err := p.BeginTransaction(ctx)
		if err != nil {
			t.Errorf("Expecting no error but found err %v", err)
		}

		if _, ok := dbtx.(*dbTx); !ok {
			t.Errorf("Expecting instance of transaction %v", err)
		}

	})
	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("Valid circuit breaker error : connection refused")

		mock.ExpectBegin().WillReturnError(errors.New("connection refused"))
		ctx := context.Background()
		_, gotErr := p.BeginTransaction(ctx)
		if gotErr == nil {
			t.Errorf("Expecting error but found nil %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}
	})
	t.Run("Invalid cb error", func(t *testing.T) {
		expectedErr := errors.New("duplicate_column")

		mock.ExpectBegin().WillReturnError(errors.New("duplicate_column"))
		ctx := context.Background()
		_, gotErr := p.BeginTransaction(ctx)
		if gotErr == nil {
			t.Errorf("Expecting error but found nil %v", gotErr)
		}

		if !reflect.DeepEqual(expectedErr, gotErr) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, gotErr)
		}
	})
}

func TestExecContext(t *testing.T) {
	t.Run("Error running exec query", func(t *testing.T) {
		p, mock, closeDB := setup(t)
		defer closeDB()

		mock.ExpectBegin()
		mock.ExpectExec(inserQuery).WillReturnError(fmt.Errorf("some error"))

		ctx := context.Background()

		dbtx, err := p.BeginTransaction(ctx)
		err = dbtx.ExecContext(ctx, inserQuery)
		if err == nil {
			t.Errorf("Expecting error but found nil")
		}
	})

	t.Run("success exec query", func(t *testing.T) {
		p, mock, closeDB := setup(t)
		defer closeDB()

		mock.ExpectBegin()
		mock.ExpectExec(updateQuery).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.Background()

		dbtx, err := p.BeginTransaction(ctx)
		if err != nil {
			t.Errorf("Expecting no error but found err %v", err)
		}
		err = dbtx.ExecContext(ctx, updateQuery)
		if err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}
	})
	t.Run("Valid cb error", func(t *testing.T) {
		expectedErr := errors.New("Valid circuit breaker error : connection refused")
		p, mock, closeDB := setupCb(t)
		mock.ExpectBegin()
		ctx := context.Background()
		dbtx, err := p.BeginTransaction(ctx)
		if err != nil {
			t.Errorf("Expecting no error but found err %v", err)
		}
		mock.ExpectExec(selectQuery).WillReturnError(errors.New("connection refused"))
		err = dbtx.ExecContext(ctx, selectQuery)
		defer closeDB()

		if err == nil {
			t.Errorf("Expecting err but found nil %v", err)
		}

		if !reflect.DeepEqual(expectedErr, err) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, err)
		}
	})
	t.Run("Invalid cb error", func(t *testing.T) {
		expectedErr := errors.New("duplicate_column")
		p, mock, closeDB := setupCb(t)
		mock.ExpectBegin()
		ctx := context.Background()
		dbtx, err := p.BeginTransaction(ctx)
		if err != nil {
			t.Errorf("Expecting no error but found err %v", err)
		}
		mock.ExpectExec(selectQuery).WillReturnError(errors.New("duplicate_column"))

		err = dbtx.ExecContext(ctx, selectQuery)
		defer closeDB()

		if err == nil {
			t.Errorf("Expecting err but found nil %v", err)
		}

		if !reflect.DeepEqual(expectedErr, err) {
			t.Errorf("Expected error: %v but got error: %v", expectedErr, err)
		}
	})
}

func TestSelectObjectsContext(t *testing.T) {
	p, mock, closeDB := setupCb(t)
	defer closeDB()

	t.Run("error", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectBegin()
		mock.ExpectQuery(selectQuery).WillReturnError(errors.New("Some Errors"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}

		dbtx, err := p.BeginTransaction(ctx)
		err = dbtx.SelectObjectsContext(ctx, &entities, selectQuery)
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
		ctx := context.Background()

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")

		mock.ExpectQuery("Select from Ticket").WillReturnRows(rows)
		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}

		dbtx, err := p.BeginTransaction(ctx)

		err = dbtx.SelectObjectsContext(ctx, &entities, "Select from Ticket")
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
	t.Run("Valid CB Error", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectBegin()
		mock.ExpectQuery(selectQuery).WillReturnError(errors.New("connection refused"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}

		dbtx, err := p.BeginTransaction(ctx)
		err = dbtx.SelectObjectsContext(ctx, &entities, selectQuery)
		if err == nil {
			t.Errorf("Expecting error but found nil %v", err)
		}

		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("InValid CB Error", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectBegin()
		mock.ExpectQuery(selectQuery).WillReturnError(errors.New("duplicate_column"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}

		dbtx, err := p.BeginTransaction(ctx)
		err = dbtx.SelectObjectsContext(ctx, &entities, selectQuery)
		if err == nil {
			t.Errorf("Expecting error but found nil %v", err)
		}

		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("success with no records", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"})
		mock.ExpectQuery("Select from Ticket").WillReturnRows(rows)

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}

		dbtx, err := p.BeginTransaction(ctx)
		err = dbtx.SelectObjectsContext(ctx, &entities, "Select from Ticket")
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
}

func TestSelectObjectWithPrepareAndProcessContext(t *testing.T) {
	p, mock, closeDB := setupCb(t)
	defer closeDB()

	var errorReceivedInCallback error
	callbackInvokeCount := 0
	sendErrorFromCallback := false
	sendValidCBError := false
	sendInvalidCBError := false
	inputQuery := "select from test"
	inputBindVariable := 1
	injectedError := errors.New("injected-error")
	cbError := errors.New("Valid circuit breaker error : connection refused")
	invalidcbError := errors.New("duplicate_column")

	mockProcessObjectCallback := func(object interface{}, err error) error {
		callbackInvokeCount++
		if errorReceivedInCallback == nil {
			errorReceivedInCallback = err
		}
		if sendErrorFromCallback {
			return injectedError
		}
		if sendValidCBError {
			errorReceivedInCallback = cbError
		}
		if sendInvalidCBError {
			errorReceivedInCallback = invalidcbError
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
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 2,
		},

		{
			scenario: "Query failed",
			setupMocks: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnError(injectedError)
			},
			expectedCallbackInvocationCount: 1,
			expectedCallbackError:           injectedError,
		},

		{
			scenario: "Row scan failed",
			setupMocks: func() {
				mock.ExpectBegin()
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
				mock.ExpectBegin()
				sendErrorFromCallback = true
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 1,
		},
		{
			scenario: "Valid Cb error",
			setupMocks: func() {
				mock.ExpectBegin()
				sendValidCBError = true
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnError(errors.New("connection refused"))
			},
			expectedCallbackInvocationCount: 1,
			expectedCallbackError:           cbError,
		},
		{
			scenario: "InValid Cb error",
			setupMocks: func() {
				mock.ExpectBegin()
				sendInvalidCBError = true
				mock.ExpectQuery(inputQuery).WithArgs(inputBindVariable).WillReturnError(invalidcbError)
			},
			expectedCallbackInvocationCount: 1,
			expectedCallbackError:           invalidcbError,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			errorReceivedInCallback = nil
			callbackInvokeCount = 0
			sendErrorFromCallback = false
			sendValidCBError = false
			sendInvalidCBError = false
			test.setupMocks()
			ctx := context.Background()
			dbtx, _ := p.BeginTransaction(ctx)

			dbtx.SelectObjectWithPrepareAndProcessContext(ctx, &entity{}, mockProcessObjectCallback, inputQuery, inputBindVariable)

			if errorReceivedInCallback != test.expectedCallbackError {
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
func TestSelectObjectAndProcessContext(t *testing.T) {
	p, mock, closeDB := setup(t)
	defer closeDB()

	var errorReceivedInCallback error
	callbackInvokeCount := 0
	sendErrorFromCallback := false
	sendValidCBError := false
	sendInvalidCBError := false
	inputQuery := "select from test"
	injectedError := errors.New("injected-error")
	cbError := errors.New("Valid circuit breaker error : connection refused")
	invalidcbError := errors.New("duplicate_column")

	mockProcessObjectCallback := func(object interface{}, err error) error {
		callbackInvokeCount++
		if errorReceivedInCallback == nil {
			errorReceivedInCallback = err
		}
		if sendErrorFromCallback {
			return injectedError
		}

		if sendValidCBError {
			errorReceivedInCallback = cbError
		}
		if sendInvalidCBError {
			errorReceivedInCallback = invalidcbError
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
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				mock.ExpectQuery(inputQuery).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 2,
		},
		{
			scenario: "Query failed",
			setupMocks: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(inputQuery).WillReturnError(injectedError)
			},
			expectedCallbackInvocationCount: 1,
			expectedCallbackError:           injectedError,
		},

		{
			scenario: "Row scan failed",
			setupMocks: func() {
				mock.ExpectBegin()
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
				mock.ExpectBegin()
				sendErrorFromCallback = true
				rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
				mock.ExpectQuery(inputQuery).WillReturnRows(rows)
			},
			expectedCallbackInvocationCount: 1,
		},
		{
			scenario: "Valid CB error",
			setupMocks: func() {
				mock.ExpectBegin()
				sendValidCBError = true
				mock.ExpectQuery(inputQuery).WillReturnError(errors.New("connection refused"))
			},
			expectedCallbackInvocationCount: 1,
			expectedCallbackError:           cbError,
		},
		{
			scenario: "InValid CB error",
			setupMocks: func() {
				mock.ExpectBegin()
				sendInvalidCBError = true
				mock.ExpectQuery(inputQuery).WillReturnError(invalidcbError)
			},
			expectedCallbackInvocationCount: 1,
			expectedCallbackError:           invalidcbError,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			errorReceivedInCallback = nil
			callbackInvokeCount = 0
			sendErrorFromCallback = false
			sendValidCBError = false
			sendInvalidCBError = false

			test.setupMocks()
			ctx := context.Background()
			dbtx, _ := p.BeginTransaction(ctx)

			dbtx.SelectObjectAndProcessContext(ctx, &entity{}, mockProcessObjectCallback, inputQuery)

			if errorReceivedInCallback != test.expectedCallbackError {
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
func TestSelectObjectsWithPrepareContext(t *testing.T) {
	p, mock, closeDB := setup(t)
	defer closeDB()

	t.Run("error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(selectPrepareQuery).WithArgs("2").WillReturnError(errors.New("Some Errors"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}
		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		err := dbtx.SelectObjectsWithPrepareContext(ctx, &entities, selectPrepareQuery, "2")
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
		rows := sqlmock.NewRows([]string{"Id", "InTime", "OutTime"}).AddRow(1, "10 Aug", "11 Aug").AddRow(2, "12 Aug", "13 Aug")
		mock.ExpectBegin()
		mock.ExpectQuery(selectPrepareQuery).WithArgs("2").WillReturnRows(rows)

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		err := dbtx.SelectObjectsWithPrepareContext(ctx, &entities, selectPrepareQuery, "2")
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
	t.Run("Valid CB Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(selectPrepareQuery).WithArgs("2").WillReturnError(errors.New("connection refused"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		goterr := dbtx.SelectObjectsWithPrepareContext(ctx, &entities, selectPrepareQuery, "2")
		if goterr == nil {
			t.Errorf("Expecting err but found nil %v", goterr)
		}

		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}

	})
	t.Run("InValid CB Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(selectPrepareQuery).WithArgs("2").WillReturnError(errors.New("duplicate_column"))

		entities := []struct {
			ID      int    `db:"Id"`
			Intime  string `db:"InTime"`
			Outtime string `db:"OutTime"`
		}{}

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		goterr := dbtx.SelectObjectsWithPrepareContext(ctx, &entities, selectPrepareQuery, "2")
		if goterr == nil {
			t.Errorf("Expecting err but found nil %v", goterr)
		}

		if len(entities) != 0 {
			t.Errorf("Expecting 0 records returned but found %d", len(entities))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}

	})

}

func TestExecWithPrepareContext(t *testing.T) {
	p, mock, closeDB := setup(t)
	defer closeDB()

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(insertPrepareQuery).WithArgs("2").WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		if err := dbtx.ExecWithPrepareContext(ctx, insertPrepareQuery, "2"); err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}

	})

	t.Run("Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(insertPrepareQuery).WithArgs("2").WillReturnError(errors.New("Some Error"))

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		if err := dbtx.ExecWithPrepareContext(ctx, insertPrepareQuery, "2"); err == nil {
			t.Errorf("Expecting error but found nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Valid CB Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(insertPrepareQuery).WithArgs("2").WillReturnError(errors.New("connection refused"))

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		if gotErr := dbtx.ExecWithPrepareContext(ctx, insertPrepareQuery, "2"); gotErr == nil {
			t.Errorf("Expecting error found nil %v", gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("InValid CB Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(insertPrepareQuery).WithArgs("2").WillReturnError(errors.New("duplicate_column"))

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		if gotErr := dbtx.ExecWithPrepareContext(ctx, insertPrepareQuery, "2"); gotErr == nil {
			t.Errorf("Expecting error found nil %v", gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRollback(t *testing.T) {
	p, mock, closeDB := setup(t)
	defer closeDB()

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectRollback()

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		if err := dbtx.Rollback(); err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Valid CB Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectRollback().WillReturnError(errors.New("connection refused"))

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		if gotErr := dbtx.Rollback(); gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("InValid CB Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectRollback().WillReturnError(errors.New("duplicate_column"))

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		if gotErr := dbtx.Rollback(); gotErr == nil {
			t.Errorf("Expecting nil but found err %v", gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestCommit(t *testing.T) {
	p, mock, closeDB := setup(t)
	defer closeDB()

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectCommit()

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		if err := dbtx.Commit(); err != nil {
			t.Errorf("Expecting nil but found err %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Valid CB Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(errors.New("connection refused"))

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		if gotErr := dbtx.Commit(); gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("InValid CB Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(errors.New("duplicate_column"))

		ctx := context.Background()
		dbtx, _ := p.BeginTransaction(ctx)

		if gotErr := dbtx.Commit(); gotErr == nil {
			t.Errorf("Expecting err but found nil %v", gotErr)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}
