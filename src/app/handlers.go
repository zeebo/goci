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
	var ws []*Work
	err := ctx.DB.C(collection).Find(nil).Sort(d{"$natural": -1}).Limit(10).All(&ws)
	if err != nil {
		internal_error(w, req, ctx, err)
		return
	}
	ctx.Set("Recent", ws)
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

func handle_build_info(w http.ResponseWriter, req *http.Request, ctx *Context) {
	id := req.FormValue(":id")
	if id == "" {
		perform_status(w, ctx, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	var wk *Work
	err := ctx.DB.C(collection).Find(d{"builds._id": id}).One(&wk)
	if err != nil {
		internal_error(w, req, ctx, err)
		return
	}
	var bd *Build
	for _, b := range wk.Builds {
		if b.ID == id {
			bd = b
			break
		}
	}
	if bd == nil {
		internal_error(w, req, ctx, fmt.Errorf("%s: queryed but not found", id))
		return
	}
	ctx.Set("Build", bd)
	base_execute(w, ctx, tmpl_root("blocks", "build.block"))
}
