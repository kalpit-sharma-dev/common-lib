package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/db"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/db/mssql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/db/postgresql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

const (
	query       = "Select * from Ticket_Tracing where TicketID = ?"
	insertQuery = "Insert into Ticket_Tracing values(?,?,?)"
	updateQuery = "Update Ticket_Tracing set OutTime = ? where TicketID = ?"
	deleteQuery = "Delete from Ticket_Tracing where TicketID = ?"

	selectQuery  = "Select * from Ticket_Tracing where TicketID = 5"
	insertQuery2 = "Insert into Ticket_Tracing values(10,GETDATE(),GETDATE())"

	postgresPreparedQuery = "SELECT * FROM public.test WHERE column_one > $1"
	postgresQuery         = "SELECT * FROM public.test WHERE column_one >"
)

func main() {

	logger.Create(logger.Config{}) //no lint
	db.Logger = logger.Get

	mssqlExample()
	postgresqlExample()

}

func mssqlExample() {
	//Get DbProvider Instance
	db, err := db.GetDbProvider(db.Config{DbName: "NOCBO",
		Server:     "10.2.27.41",
		Password:   "its",
		UserID:     "its",
		Driver:     mssql.Dialect,
		CacheLimit: 200})

	if err != nil {
		fmt.Println(err)
		return
	}

	transactionExec(db)
	transactionSelect(db)
	transactionSingleConnection(db)

	//Select Query by Creating Prepared Statement
	rows, err := db.SelectWithPrepare(query, 1)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(rows)

	//Insert Query by Creating Prepared Statement
	err = db.ExecWithPrepare(insertQuery, 5, time.Now(), time.Now())
	if err != nil {
		fmt.Println(err)
		return
	}

	//Update Query by Creating Prepared Statement
	err = db.ExecWithPrepare(updateQuery, time.Now(), 1)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Delete Query by Creating Prepared Statement
	err = db.ExecWithPrepare(deleteQuery, 5)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Close prepared statement
	err = db.CloseStatement(query)

	//Plain text select query
	rows, err = db.Select(selectQuery)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(rows)

	//Plain text exec query
	err = db.Exec(insertQuery2)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Plaintext select query with callback function to read rows
	db.SelectAndProcess(selectQuery, processRowCallback)

	//Select query using prepared statement and callback function to read rows
	db.SelectWithPrepareAndProcess(query, processRowCallback, 1)

	//examples for doing multiple calls on a single connection
	singleConnectionExample(db)
}

// Callback func to process table row
func processRowCallback(row db.Row) {
	fmt.Printf("Row: %v | Error: %v", row.Columns, row.Error)
}

func postgresqlExample() {
	additionalConfig := make(map[string]string)
	additionalConfig[postgresql.ServerPortKey] = "5432"
	additionalConfig[postgresql.SSLModeKey] = "disable"

	//Get DbProvider Instance
	db, err := db.GetDbProvider(db.Config{DbName: "postgres",
		Server:           "localhost",
		Password:         "admin",
		UserID:           "postgres",
		Driver:           postgresql.Dialect,
		AdditionalConfig: additionalConfig,
		CacheLimit:       200})

	if err != nil {
		//handle the error
		fmt.Printf("error occured during connection : %+v\n", err)
		return
	}

	//examples are for prepared statements
	preparedStatementExample(db)

	//examples are for SQL query
	queryExample(db)

	//examples for doing multiple calls on a single connection
	singleConnectionExample(db)

}

func preparedStatementExample(db db.DatabaseProvider) {
	//generate/use existing transactionID for each new application transaction (e.g. new request, new scheduler run, etc.)
	transactionID := "abcd-1"

	if err := db.PrepareStatement(transactionID, postgresPreparedQuery); err != nil {
		//handle the error
		fmt.Printf("error occured while preparing statement: %+v\n", err)
		return
	}

	//Approach 1: select all rows at once by passing pointer to slice of struct object
	records := []testRecord{}
	err := db.SelectObjectsWithPrepare(transactionID, &records, postgresPreparedQuery, 1)
	if err != nil {
		//handle the error
		fmt.Printf("error occured while selecting using prepared statement: %+v\n", err)
		return
	}
	//all rows should be available in the slice
	for i, record := range records {
		fmt.Printf("Record number: %d, values: %d, %d, %s\n", i, record.ColumnOne, record.ColumnTwo, record.ColumnThree)
	}

	//Approach 2: select row one by one by passing pointer to single struct object and get callback with populated object
	record := testRecord{}
	db.SelectObjectWithPrepareAndProcess(transactionID, &record, callback, postgresPreparedQuery, 1)

}

func queryExample(db db.DatabaseProvider) {
	//generate/use existing transactionID for each new application transaction (e.g. new request, new scheduler run, etc.)
	transactionID := "abcd-2"

	//Approach 1: select all rows at once by passing pointer to slice of struct object
	records := []testRecord{}
	err := db.SelectObjects(transactionID, &records, strings.Join([]string{postgresQuery, "1"}, " "))
	if err != nil {
		//handle the error
		fmt.Printf("error occured while selecting using query: %+v\n", err)
		return
	}
	//all rows should be available in the slice
	for i, record := range records {
		fmt.Printf("Record number: %d, values: %d, %d, %s\n", i, record.ColumnOne, record.ColumnTwo, record.ColumnThree)
	}

	//Approach 2: select row one by one by passing pointer to single struct object and get callback with populated object
	record := testRecord{}
	db.SelectObjectAndProcess(transactionID, &record, callback, strings.Join([]string{postgresQuery, "1"}, " "))

}

func singleConnectionExample(db db.DatabaseProvider) {
	//generate/use existing transactionID for each new application transaction (e.g. new request, new scheduler run, etc.)
	transactionID := "abcd-2"

	//Grab the connection from the pool
	ctx := context.Background()
	conn, err := db.GetSingleConnectionProvider(ctx, transactionID)
	if err != nil {
		//handle the error
		fmt.Printf("error occured while selecting using query: %+v\n", err)
		return
	}
	//Return the connection to the pool after the connection
	defer conn.Close(transactionID)

	//Call anything you could with a normal DatabaseProvider
	queryExample(conn)

}

type testRecord struct {
	ColumnOne   int    `db:"column_one"`
	ColumnTwo   int    `db:"column_two"`
	ColumnThree string `db:"column_three"`
}

var callback db.ProcessObject = processObjectCallback

func processObjectCallback(object interface{}, err error) error {
	if err != nil {
		//error occured while reading the data from db. Handle the error
		fmt.Printf("error received : %+v\n", err)
	}

	record, ok := object.(*testRecord)
	if !ok {
		//this means that wrong struct pointer was sent in the query. This typically should not happen
		fmt.Printf("wrong struct pointer\n")
	}

	//process the record
	fmt.Printf("Callback called at %v - Record values: %d, %d, %s\n", time.Now(), record.ColumnOne, record.ColumnTwo, record.ColumnThree)

	//return error only if some unrecoverable error occured and no more records are desired and reading from database should stop, otherwise return nil
	return nil
}
