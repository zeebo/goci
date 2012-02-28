package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func debugDatabase(w http.ResponseWriter, req *http.Request) {
	tid := req.URL.Query().Get(":id")
	if tid == "list" {
		dumpIds(w)
		return
	}

	id, err := strconv.ParseInt(tid, 10, 32)
	if err != nil {
		fmt.Fprintln(w, "404 page not found")
		w.WriteHeader(http.StatusNotFound)
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

func dumpIds(w io.Writer) {
	rows, err := databaseConn.Query(`SELECT object_id
		FROM objects WHERE
		    object_deleted = false`)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	var id int
	for rows.Next() {
		if err = rows.Scan(&id); err != nil {
			return
		}

		fmt.Fprintln(w, id)
	}
}
