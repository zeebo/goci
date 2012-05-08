package main

import (
	"net/http"
	"thegoods.biz/httpbuf"
)

type handlerFunc func(http.ResponseWriter, *http.Request, *Context)

func (h handlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//create a context for the handler
	ctx := init_context(req)
	defer ctx.Close()

	//set up buffering on the response
	var buf httpbuf.Buffer
	defer buf.Apply(w)

	//run the handler
	h(&buf, req, ctx)
}

func handle(pattern string, handler handlerFunc) {
	http.Handle(pattern, handler)
}
