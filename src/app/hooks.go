package main

import (
	"github"
	"log"
	"net/http"
	"strings"
)

func handle_github_hook_package(w http.ResponseWriter, req *http.Request, ctx *Context) {
	var m github.HookMessage
	body := strings.NewReader(req.FormValue("payload"))
	if err := m.Load(body); err != nil {
		log.Println("error loading hook message from github:", err)
		perform_status(w, ctx, http.StatusInternalServerError)
		return
	}
	m.Workspace = false
	work_queue <- &m
}

func handle_github_hook_workspace(w http.ResponseWriter, req *http.Request, ctx *Context) {
	var m github.HookMessage
	body := strings.NewReader(req.FormValue("payload"))
	if err := m.Load(body); err != nil {
		log.Println("error loading hook message from github:", err)
		perform_status(w, ctx, http.StatusInternalServerError)
		return
	}
	m.Workspace = true
	work_queue <- &m
}
