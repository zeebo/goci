package main

import (
	"fmt"
	"net/http"
)

//our basic handle index that demonstrates how to get data from the context
//inside a template
func handle_index(w http.ResponseWriter, req *http.Request, ctx *Context) {
	if req.URL.Path != "/" {
		perform_status(w, ctx, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-type", "text/html")
	base_execute(w, ctx, tmpl_root("blocks", "index.block"))
}

func handle_work_status(w http.ResponseWriter, req *http.Request, ctx *Context) {
	id := req.FormValue(":id")
	if id == "" {
		perform_status(w, ctx, http.StatusNotFound)
		return
	}
	st, err := work_status(ctx.DB, id)
	if err != nil {
		internal_error(w, req, ctx, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, st)
}
