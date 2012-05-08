package main

import (
	"log"
	"net/http"
	"path"
	"strconv"
)

//our basic handle index that demonstrates how to get data from the context
//inside a template
func handle_index(w http.ResponseWriter, req *http.Request, ctx *Context) {
	if req.URL.Path != "/" {
		perform_status(w, req, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-type", "text/html")
	ctx.Set("content", "Content!")
	execute(w, ctx, tmpl_root("blocks", "index.block"))
}

//make a silly handler for testing statuses
func handle_status(w http.ResponseWriter, req *http.Request, ctx *Context) {
	status, err := strconv.ParseInt(path.Base(req.URL.Path), 10, 32)
	if err != nil {
		log.Println(err)
		perform_status(w, req, http.StatusNotFound)
		return
	}
	perform_status(w, req, int(status))
}
