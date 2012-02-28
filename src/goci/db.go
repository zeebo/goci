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

func init() {
	//connect to the database
	databaseConn, err := sql.Open("postgres", os.Getenv("HEROKU_SHARED_POSTGRESQL_ROSE_URL"))
	if err != nil {
		panic(err)
	}
	//boostrap fdb
	if err := fdb.Bootstrap(databaseConn); err != nil {
		panic(err)
	}
	//run our insert loob
	go resultInsert()
}

func resultInsert() {
	for result := range resultsChan {
		if err := fdb.Update(&result); err != nil {
			errLogger.Println("Result insert:", err)
		}
	}
}
