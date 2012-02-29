package main

import (
	"database/sql"
	"github.com/zeebo/fdb"
	_ "github.com/zeebo/pq.go"
	"os"
)

var (
	//make our channel to send results in
	resultsChan  = make(chan Result)
	databaseConn *sql.DB
)

func setupDatabase() (err error) {
	//connect to the database
	databaseConn, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return
	}

	//boostrap fdb
	return fdb.Bootstrap(databaseConn)
}

func resultInsert() {
	for result := range resultsChan {
		if err := fdb.Update(&result); err != nil {
			errLogger.Println("Result insert:", err)
		}
	}
}
