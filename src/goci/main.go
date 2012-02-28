package main

import (
	"database/sql"
	"github.com/bmizerany/pat.go"
	"github.com/zeebo/fdb"
	_ "github.com/zeebo/pq.go"
	"net/http"
	"os"
)

var (
	//make our channel to send results in
	resultsChan  = make(chan Result)
	databaseConn *sql.DB
)

func main() {
	var err error
	//connect to the database
	databaseConn, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	//boostrap fdb
	if err = fdb.Bootstrap(databaseConn); err != nil {
		panic(err)
	}
	//run our insert loob
	go resultInsert()

	//set up our routing system
	m := pat.New()
	m.Get("/debug/:id", http.HandlerFunc(debugDatabase))
	http.Handle("/debug/", m)
	http.HandleFunc("/push", handlePush)

	//listen
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		errLogger.Fatal(err)
	}
}

func resultInsert() {
	for result := range resultsChan {
		if err := fdb.Update(&result); err != nil {
			errLogger.Println("Result insert:", err)
		}
	}
}
