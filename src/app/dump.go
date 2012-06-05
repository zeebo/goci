package main

import (
	"github.com/zeebo/pretty"
	"net/http"
)

func handle_dump(w http.ResponseWriter, req *http.Request, ctx *Context) {
	w.Header().Set("Content-Type", "text/html")
	id, coll := req.FormValue(":id"), req.FormValue(":coll")

	var (
		x   interface{}
		err error
	)

	switch coll {
	case "Work":
		var w Work
		err = load(coll, d{"_id": id}, &w)
		x = w
	case "Test":
		var t Test
		err = load(coll, d{"_id": id}, &t)
		x = t
	case "Build":
		var b Build
		err = load(coll, d{"_id": id}, &b)
		x = b
	default:
		perform_status(w, ctx, http.StatusNotFound)
		return
	}

	if err != nil {
		internal_error(w, req, ctx, err)
		return
	}
	ctx.Set("Item", pretty.Formatter(x))
	base_execute(w, ctx, tmpl_root("blocks", "dump.block"))
}
