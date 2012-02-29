package main

import (
	"database/sql"
	_ "github.com/zeebo/pq.go"
)

type ResultRingBuf struct {
	data   []*Result
	size   int
	head   int
	filled bool
}

func NewResultRingBuf(size int) *ResultRingBuf {
	return &ResultRingBuf{
		data: make([]*Result, 2*size),
		size: size,
	}
}

func (d *ResultRingBuf) Push(val *Result) {
	d.data[d.head], d.data[d.head+d.size] = val, val
	d.head += 1

	if d.head >= d.size {
		d.head = 0
		d.filled = true
	}
}

func (d *ResultRingBuf) Slice() []*Result {
	if d.filled {
		return d.data[d.head : d.head+d.size]
	}
	return d.data[:d.head]
}

var (
	//make our channel to send results in
	resultsChan  = make(chan Result)
	databaseConn *sql.DB

	recentResults = make(chan []*Result)
)

func setupDatabase() (err error) {
	//connect to the database

	/*
		databaseConn, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
		if err != nil {
			return
		}

		//boostrap fdb
		return fdb.Bootstrap(databaseConn)
	*/

	//for now just do an in memory store
	return
}

func resultInsert() {
	buf := NewResultRingBuf(100) //store last 100 results
	for {
		select {
		case result := <-resultsChan:
			buf.Push(&result)
		case recentResults <- buf.Slice():
		}
	}

	/*
		for result := range resultsChan {
			if err := fdb.Update(&result); err != nil {
				errLogger.Println("Result insert:", err)
			}
		}
	*/
}
