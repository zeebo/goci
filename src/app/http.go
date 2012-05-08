package main

import "net/http"

type handlerFunc func(http.ResponseWriter, *http.Request, *Context)

func (h handlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := init_context(req)
	defer ctx.Close()
	h(w, req, ctx)
}

func wrap_func(h handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := init_context(req)
		defer ctx.Close()
		h(w, req, ctx)
	}
}

func handle_func(pattern string, handler handlerFunc) {
	http.HandleFunc(pattern, wrap_func(handler))
}

func handle(pattern string, handler handlerFunc) {
	http.Handle(pattern, handler)
}
