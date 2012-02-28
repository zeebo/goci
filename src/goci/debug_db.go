package main

import (
	"fmt"
	"github.com/bmizerany/pat.go"
	"net/http"
	"strconv"
)

func init() {
	m := pat.New()
	m.Get("/:id", http.HandlerFunc(debugDatabase))
	http.Handle("/debug/", m)
}

func debugDatabase(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 32)
	if err != nil {
		fmt.Fprintln(w, "Invalid id")
		return
	}
	rows, err := databaseConn.Query(`SELECT attribute_key, attribute_value
		FROM attributes WHERE
		    object_id = $1
		AND attribute_archived = false
		AND attribute_preview = false`, id)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	fmt.Fprintln(w, "<pre>")

	//loop over them
	var key string
	var value []byte
	for rows.Next() {
		if err = rows.Scan(&key, &value); err != nil {
			return
		}

		fmt.Fprintln(w, key, string(value))
	}
}
