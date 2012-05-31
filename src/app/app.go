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

//run the basic setup needed to serve web pages
func init() {
	//set our compiler mode
	tmplmgr.CompileMode(mode)

	//add blocks to base template
	base_template.Blocks(tmpl_root("*.block"))

	go run_setup()
}

func main() {
	handleGet("/bins/{id}", handlerFunc(handle_test_request), "test_request")
	handlePost("/bins/{id}", handlerFunc(handle_test_response), "test_response")

	handlePost("/hooks/github", handlerFunc(handle_github_hook), "github_hook")

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
