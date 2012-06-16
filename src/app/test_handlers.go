package main

import (
	"log"
	"net/http"
	"worker"
)

func handle_test_request(w http.ResponseWriter, req *http.Request, ctx *Context) {
	id := req.FormValue(":id")
	err := worker.Serve(w, id)
	if err != nil {
		log.Printf("serve %s: %s", id, err)
	}
}

func handle_test_response(w http.ResponseWriter, req *http.Request, ctx *Context) {
	id := req.URL.Query().Get(":id")
	err := worker.Response(req.Body, id)
	if err != nil {
		log.Printf("response %s: %s", id, err)
	}
}

func handle_test_error(w http.ResponseWriter, req *http.Request, ctx *Context) {
	id := req.URL.Query().Get(":id")
	err := worker.Error(req.Body, id)
	if err != nil {
		log.Printf("error %s: %s", id, err)
	}
}
