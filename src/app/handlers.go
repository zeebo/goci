package main

import "net/http"

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
