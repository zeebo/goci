package main

import (
	"bitbucket"
	"builder"
	"github"
	"io"
	"log"
	"net/http"
	"strings"
)

type work_loader interface {
	builder.Work
	Load(io.Reader) error
}

func perform_hook(w http.ResponseWriter, req *http.Request, ctx *Context, l work_loader) {
	body := strings.NewReader(req.FormValue("payload"))
	if err := l.Load(body); err != nil {
		log.Println("error loading hook message from github:", err)
		perform_status(w, ctx, http.StatusInternalServerError)
		return
	}
	work_queue <- l
}

func handle_github_hook_package(w http.ResponseWriter, req *http.Request, ctx *Context) {
	m := &github.HookMessage{Workspace: false}
	perform_hook(w, req, ctx, m)
}

func handle_github_hook_workspace(w http.ResponseWriter, req *http.Request, ctx *Context) {
	m := &github.HookMessage{Workspace: true}
	perform_hook(w, req, ctx, m)
}

func handle_bitbucket_hook_package(w http.ResponseWriter, req *http.Request, ctx *Context) {
	m := &bitbucket.HookMessage{Workspace: false}
	perform_hook(w, req, ctx, m)
}

func handle_bitbucket_hook_workspace(w http.ResponseWriter, req *http.Request, ctx *Context) {
	m := &bitbucket.HookMessage{Workspace: true}
	perform_hook(w, req, ctx, m)
}
