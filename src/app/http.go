package main

import (
	"fmt"
	"net/http"
	"thegoods.biz/httpbuf"
)

type runner interface {
	Run(http.ResponseWriter, *http.Request, *Context)
}

func perform(r runner, w http.ResponseWriter, req *http.Request) {
	ctx := init_context(req)
	defer ctx.Close()

	var buf httpbuf.Buffer
	defer buf.Apply(w)

	r.Run(&buf, req, ctx)
	ctx.Save(req, w)
}

type handlerFunc func(http.ResponseWriter, *http.Request, *Context)

func (h handlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	perform(h, w, req)
}

func (h handlerFunc) Run(w http.ResponseWriter, req *http.Request, ctx *Context) {
	h(w, req, ctx)
}

type auth_func func(*http.Request, *Context) bool

type dispatcher struct {
	GET, POST, HEAD handlerFunc
	auth            auth_func
}

func (d *dispatcher) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	perform(d, w, req)
}

func (d *dispatcher) Run(w http.ResponseWriter, req *http.Request, ctx *Context) {
	if d.auth != nil && !d.auth(req, ctx) {
		perform_status(w, ctx, http.StatusNotFound)
		return
	}

	switch {
	case req.Method == "GET" && d.GET != nil:
		d.GET(w, req, ctx)
	case req.Method == "POST" && d.POST != nil:
		d.POST(w, req, ctx)
	case req.Method == "HEAD" && d.HEAD != nil:
		d.HEAD(w, req, ctx)
	default:
		perform_status(w, ctx, http.StatusNotFound)
	}
}

func wrap_runner(r runner) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		perform(r, w, req)
	})
}

func handleGet(pattern string, handler http.Handler, name string) {
	router.Add("GET", pattern, handler).Name(name)
}

func handlePost(pattern string, handler http.Handler, name string) {
	router.Add("POST", pattern, handler).Name(name)
}

func handleRequest(pattern string, handler http.Handler, name string) {
	handleGet(pattern, handler, name)
	handlePost(pattern, handler, name)
}

func reverse(name string, things ...interface{}) string {
	strs := make([]string, len(things))
	for i, th := range things {
		strs[i] = fmt.Sprint(th)
	}
	u, err := router.GetRoute(name).URL(strs...)
	if err != nil {
		panic(err)
	}
	return u.Path
}
