package main

import (
	"database/sql"
	"encoding/json"
	"github.com/zeebo/fdb"
	_ "github.com/zeebo/pq.go"
	"os"
	"time"
)

type Result struct {
	ID       int
	Repo     string
	Duration time.Duration
	Build    Status
	Test     Status
}

type Status struct {
	Passed bool
	Output string
	Error  string
}

//Store status as a json object in the database
func (s *Status) Unserialize(p []byte) error { return json.Unmarshal(p, s) }
func (s *Status) Serialize() (p []byte)      { p, _ = json.Marshal(s); return }

//assert Status is a serializer
var _ fdb.Serializer = &Status{}

var (
	//make our channel to send results in
	resultsChan  = make(chan Result)
	databaseConn *sql.DB
)

func init() {
	//connect to the database
	databaseConn, err := sql.Open("pstgres", os.Getenv("HEROKU_SHARED_POSTGRESQL_ROSE_URL"))
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
			errorLogger.Println("Result insert:", err)
		}
	}
}
