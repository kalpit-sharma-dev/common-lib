package main

import (
	"fmt"
	goLog "log"

	db "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

const (
	query       = `SELECT * FROM cats WHERE "ID" = ?`
	insertQuery = `INSERT INTO cats("ID", "Name", "Age", "Weight", "Owners") VALUES(?,?,?,?,?)`
	updateQuery = `UPDATE cats SET "Name" = ? WHERE "ID" = ? AND "Age" = ?`
	deleteQuery = `DELETE FROM cats WHERE "ID" = ? AND "Age" = ?`

	selectQuery  = `SELECT * FROM cats WHERE "ID" = 4c3766db-94f0-11e8-b068-080027f00fcc`
	insertQuery2 = `INSERT INTO cats("ID", "Name", "Age", "Weight", "Owners") VALUES(9f779d2c-4de1-4c93-ae61-7c1e3d043653, 'Tommy', 3, 2, ['Molly'])`

	catAge     = 3
	sampleUUID = "9f779d2c-4de1-4c93-ae61-7c1e3d043653"
)

func main() {
	log, err := logger.Create(logger.Config{
		FileName: "./platform-query-sample.log",
		LogLevel: logger.DEBUG,
	})
	if err != nil {
		panic(err)
	}

	cassandraHosts := []string{"localhost:9042"} // replace cassandra with your host
	cassandraKeyspace := "platform_pets_db"
	cassandraTimeout := "3s"

	err = db.Load(cassandraHosts, cassandraKeyspace, cassandraTimeout, log)
	if err != nil {
		panic(err)
	}

	// Select Query by Creating Prepared Statement
	data, err := db.Session.Select(query, "4c3766db-94f0-11e8-b068-080027f00fcc")
	if err != nil {
		goLog.Print(err)
		return
	}
	fmt.Println("Select with prepare: ", data)

	// Insert Query by Creating Prepared Statement
	err = db.Session.Exec(insertQuery, sampleUUID, "Tommy", catAge, 2, []string{"Molly"})
	if err != nil {
		goLog.Print(err)
		return
	}

	// Update Query by Creating Prepared Statement
	err = db.Session.Exec(updateQuery, "Jerry", sampleUUID, catAge)
	if err != nil {
		goLog.Print(err)
		return
	}

	// Delete Query by Creating Prepared Statement
	err = db.Session.Exec(deleteQuery, sampleUUID, catAge)
	if err != nil {
		goLog.Print(err)
		return
	}

	data, err = db.Session.Select(selectQuery)
	if err != nil {
		goLog.Print(err)
		return
	}
	fmt.Println("Select: ", data)

	err = db.Session.Exec(insertQuery2)
	if err != nil {
		goLog.Print(err)
		return
	}
}
