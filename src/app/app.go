package main

import (
	"code.google.com/p/gorilla/pat"
	"code.google.com/p/gorilla/sessions"
	"log"
	"net/http"
	"path/filepath"
	"thegoods.biz/tmplmgr"
)

const (
	appname   = "goci"
	store_key = "foobar"
)

var (
	mode          = tmplmgr.Development
	assets_dir    = filepath.Join(env("APPROOT", ""), "assets")
	template_dir  = filepath.Join(env("APPROOT", ""), "templates")
	dist_dir      = filepath.Join(env("APPROOT", ""), "dist")
	base_template = tmplmgr.Parse(tmpl_root("base.tmpl"))
	store         = sessions.NewCookieStore([]byte(store_key))
	base_meta     = &Meta{
		CSS: list{
			"bootstrap.min.css",
			"bootstrap-responsive.min.css",
			"main.css",
		},
		JS: list{
			"jquery.min.js",
			"jquery-ui.min.js",
			"bootstrap.js",
		},
		BaseTitle: appname,
	}

	router = pat.New()
)

func main() {
	//set our compiler mode
	tmplmgr.CompileMode(mode)

	//add blocks to base template
	base_template.Blocks(tmpl_root("*.block"))

	//set up the environment which kicks off the work queue
	go run_setup()

	//spawn our service goroutines
	go run_test_scheduler()
	go run_run_scheduler()
	go run_saver()

	//set up our handlers
	handleGet("/bins/{id}", handlerFunc(handle_test_request), "test_request")
	handlePost("/bins/{id}/err", handlerFunc(handle_test_error), "test_error") //more specific one has to be listed first
	handlePost("/bins/{id}", handlerFunc(handle_test_response), "test_response")

	handlePost("/hooks/github/package", handlerFunc(handle_github_hook_package), "github_hook_package")
	handlePost("/hooks/github/workspace", handlerFunc(handle_github_hook_workspace), "github_hook_workspace")

	handleRequest("/foo", handlerFunc(handle_simple_work), "foo")

	//add our index with 404 support
	handleRequest("/", handlerFunc(handle_index), "index")

	//set up our router
	http.Handle("/", router)
	serve_static("/assets", asset_root(""))
	if err := http.ListenAndServe(":"+env("PORT", "9080"), nil); err != nil {
		log.Fatal(err)
	}
}
