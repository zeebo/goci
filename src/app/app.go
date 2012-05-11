package main

import (
	"builder"
	"code.google.com/p/gorilla/sessions"
	"log"
	"net/http"
	"path/filepath"
	"setup"
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

	GOROOT  = "/usr/local/go"
	VCSPATH = "/usr/local/bin"
	PATH    = "/usr/bin:/usr/local/bin:/usr/local/go/bin"
)

func init() {
	//set our compiler mode
	tmplmgr.CompileMode(mode)

	//add blocks to base template
	base_template.Blocks(tmpl_root("*.block"))

	//set the GOROOT and PATH for the builder
	builder.GOROOT = GOROOT
	builder.PATH = PATH

	//ensure we have the go tool and vcs
	if err := setup.EnsureTool(GOROOT); err != nil {
		log.Fatal(err)
	}
	if err := setup.EnsureVCS(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	handle("/", handle_index)
	handle("/status/", handle_status)
	serve_static("/assets", asset_root(""))
	if err := http.ListenAndServe(":"+env("PORT", "9080"), nil); err != nil {
		log.Fatal(err)
	}
}
