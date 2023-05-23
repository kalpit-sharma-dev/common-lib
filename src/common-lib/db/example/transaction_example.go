package main

import (
	"context"
	"fmt"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/db"
)

func transactionExec(db db.DatabaseProvider) {
	// Start transaction on db provider
	ctx := context.Background()
	tx, err := db.BeginTransaction(ctx)
	if err != nil {
		fmt.Println("Error starting transaction")
		return
	}

	// Execute insert query with prepared statement in transaction
	err = tx.ExecWithPrepareContext(ctx, insertQuery, 5, time.Now(), time.Now())
	if err != nil {
		fmt.Println(err)
	}

	// Execute plaintext insert query in transaction
	err = tx.ExecContext(ctx, insertQuery2)
	if err != nil {
		fmt.Println(err)
	}

	// Rollback transaction
	tx.Rollback()
}

// Record ...
type Record struct {
	TicketID int    `db:"TicketID"`
	InTime   string `db:"InTime"`
	OutTime  string `db:"OutTime"`
}

func transactionSelect(db db.DatabaseProvider) {
	// Select queries in transaction
	ctx := context.Background()
	tx, _ := db.BeginTransaction(ctx)

	//Approach 1: select all rows at once by passing pointer to slice of struct object
	records := []Record{}
	err := tx.SelectObjectsContext(ctx, &records, selectQuery)
	if err != nil {
		//handle the error
		fmt.Printf("error occured while selecting using query: %+v\n", err)
		return
	}
	//all rows should be available in the slice
	for i, record := range records {
		fmt.Printf("Record number: %d, values: %d, %s, %s\n", i, record.TicketID, record.InTime, record.OutTime)
	}

	//Approach 2: select row one by one by passing pointer to single struct object and get callback with populated object
	record := Record{}
	tx.SelectObjectAndProcessContext(ctx, &record, callbackTransaction, selectQuery)

	//Approach 3: select all rows at once by passing pointer to slice of struct object
	err = tx.SelectObjectsWithPrepareContext(ctx, &records, query, 5)
	if err != nil {
		//handle the error
		fmt.Printf("error occured while selecting using prepared statement: %+v\n", err)
		return
	}
	//all rows should be available in the slice
	for i, record := range records {
		fmt.Printf("Record number: %d, values: %d, %s, %s\n", i, record.TicketID, record.InTime, record.OutTime)
	}

	//Approach 4: select row one by one by passing pointer to single struct object and get callback with populated object
	tx.SelectObjectWithPrepareAndProcessContext(ctx, &record, callbackTransaction, query, 5)

	// Commit transaction
	tx.Commit()
}

var callbackTransaction db.ProcessObject = processObjectCallbackTransaction

func processObjectCallbackTransaction(object interface{}, err error) error {
	if err != nil {
		//error occured while reading the data from db. Handle the error
		fmt.Printf("error received : %+v\n", err)
	}

	record, ok := object.(*Record)
	if !ok {
		//this means that wrong struct pointer was sent in the query. This typically should not happen
		fmt.Printf("wrong struct pointer\n")
	}

	//process the record
	fmt.Printf("Callback called at %v - Record values: %d, %s, %s\n", time.Now(), record.TicketID, record.InTime, record.OutTime)

	//return error only if some unrecoverable error occured and no more records are desired and reading from database should stop, otherwise return nil
	return nil
}

func transactionSingleConnection(db db.DatabaseProvider) {
	// Transaction on single connection provider
	ctx := context.Background()
	conn, err := db.GetSingleConnectionProvider(ctx, "")
	defer conn.Close("")

	tx, err := conn.BeginTransaction(ctx)
	if err != nil {
		fmt.Println("Error starting transaction")
		return
	}

	err = tx.ExecWithPrepareContext(ctx, insertQuery, 5, time.Now(), time.Now())
	if err != nil {
		fmt.Println(err)
	}

	err = tx.ExecContext(ctx, insertQuery2)
	if err != nil {
		fmt.Println(err)
	}

	tx.Rollback()
}
