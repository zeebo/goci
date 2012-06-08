package main

import (
	"code.google.com/p/gorilla/pat"
	"code.google.com/p/gorilla/sessions"
	"heroku"
	"launchpad.net/mgo"
	"log"
	"net/http"
	"net/url"
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
			"bootstrap-superhero.min.css",
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
	router  = pat.New()
	db_name = appname
	db      *mgo.Database
	hclient *heroku.Client
)

func main() {
	//set up our heroku client
	hclient = heroku.New(need_env("APPNAME"), need_env("APIKEY"))

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

	//connect to mongo
	db_path := "localhost"
	if conf := env("MONGOLAB_URI", ""); conf != "" {
		db_path = conf
		parsed, err := url.Parse(conf)
		if err != nil {
			log.Fatalf("Error parsing MONGOLAB_URI: %q: %s", conf, err)
		}
		db_name = parsed.Path[1:]
	}
	log.Printf("\tdb_path: %s\n\tdb_name: %s", db_path, db_name)

	db_sess, err := mgo.Dial(db_path)
	if err != nil {
		log.Fatalf("error connecting to database: %s", err)
	}
	db_sess.SetMode(mgo.Strong, true)
	db = db_sess.DB(db_name)

	//set up our handlers
	handleGet("/bins/{id}", handlerFunc(handle_test_request), "test_request")
	handlePost("/bins/{id}/err", handlerFunc(handle_test_error), "test_error") //more specific one has to be listed first
	handlePost("/bins/{id}", handlerFunc(handle_test_response), "test_response")

	handlePost("/hooks/github/package", handlerFunc(handle_github_hook_package), "github_hook_package")
	handlePost("/hooks/github/workspace", handlerFunc(handle_github_hook_workspace), "github_hook_workspace")

	handleRequest("/foo", handlerFunc(handle_simple_work), "foo")
	handleRequest("/dump/{coll}/{id}", handlerFunc(handle_dump), "dump")

	//add our index with 404 support
	handleRequest("/", handlerFunc(handle_index), "index")

	//set up our router
	http.Handle("/", router)
	serve_static("/assets", asset_root(""))
	if err := http.ListenAndServe(":"+env("PORT", "9080"), nil); err != nil {
		log.Fatal(err)
	}
}
