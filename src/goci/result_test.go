package main

import (
	"database/sql"
	"github.com/zeebo/fdb"
	_ "github.com/zeebo/pq.go"
	"reflect"
	"testing"
	"time"
)

func TestFDBResult(t *testing.T) {
	//only run this test on my local machine (or any darwin machine actually)
	if goHost != `darwin-amd64` {
		return
	}

	conn, err := sql.Open("postgres", "postgres://okco:@localhost:5432/okcoerrors")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	w, err := fdb.New(conn, false)
	if err != nil {
		t.Fatal(err)
	}

	r := Result{
		Repo:     "git://github.com/zeebo/heroku-basic-app.git",
		Duration: 3 * time.Second,
		Build: Status{
			Passed: true,
			Output: "some output",
			Error:  "",
		},
		Test: Status{
			Passed: false,
			Output: "some other output",
			Error:  "some error",
		},
	}

	if err := w.Update(&r); err != nil {
		t.Fatal(err)
	}

	var x Result
	if err := w.Load(&x, r.ID); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(r, x) {
		t.Fatalf("Expected %+v Got %+v", r, x)
	}
}
