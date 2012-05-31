package main

import (
	"github"
	"log"
	"net/http"
	"strings"
)

//our basic handle index that demonstrates how to get data from the context
//inside a template
func handle_github_hook(w http.ResponseWriter, req *http.Request, ctx *Context) {
	var m github.HookMessage
	body := strings.NewReader(req.FormValue("payload"))
	if err := m.Load(body); err != nil {
		log.Println("error loading hook message from github:", err)
		perform_status(w, ctx, http.StatusInternalServerError)
		return
	}
	work_queue <- &m
}