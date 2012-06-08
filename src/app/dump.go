package main

import (
	"github.com/zeebo/pretty"
	"net/http"
)

func handle_dump(w http.ResponseWriter, req *http.Request, ctx *Context) {
	w.Header().Set("Content-Type", "text/html")
	id := req.FormValue(":id")

	work, err := load(d{"_id": id})
	if err != nil {
		internal_error(w, req, ctx, err)
		return
	}

	ctx.Set("Item", pretty.Formatter(work))
	base_execute(w, ctx, tmpl_root("blocks", "dump.block"))
}
