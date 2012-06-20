package main

import (
	"code.google.com/p/gorilla/pat"
	"code.google.com/p/gorilla/sessions"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"thegoods.biz/tmplmgr"
	"worker"
)

const (
	appname   = "goci"
	store_key = "foobar"
)

var (
	mode = tmplmgr.Production

	//TODO: make sure these things happen after the env import
	assets_dir      = filepath.Join(env("APPROOT", ""), "assets")
	template_dir    = filepath.Join(env("APPROOT", ""), "templates")
	dist_dir        = filepath.Join(env("APPROOT", ""), "dist")
	base_template   = tmplmgr.Parse(tmpl_root("base.tmpl"))
	recent_template = tmplmgr.Parse(tmpl_root("recent.tmpl"))

	store     = sessions.NewCookieStore([]byte(store_key))
	base_meta = &Meta{
		CSS: list{
			"bootstrap-superhero.min.css",
			"bootstrap-responsive.min.css",
			"main.css",
		},
		JS: list{
			"jquery.min.js",
			"jquery-ui.min.js",
			"bootstrap.js",
			"recent.js",
		},
		BaseTitle: "GoCI",
	}
	router = pat.New()
)

func main() {
	//revert to development mode if debug is set
	if env("DEBUG", "") != "" {
		mode = tmplmgr.Development
	}

	//set our compiler mode
	tmplmgr.CompileMode(mode)

	//add blocks to base template
	base_template.Blocks(tmpl_root("*.block"))
	base_template.Call("reverse", reverse)
	recent_template.Blocks(tmpl_root("blocks", "recent.block"))
	recent_template.Call("reverse", reverse)

	//get our mongo credentials
	var db_name, db_path = appname, "localhost"
	if conf := env("MONGOLAB_URI", ""); conf != "" {
		db_path = conf
		parsed, err := url.Parse(conf)
		if err != nil {
			log.Fatal("Error parsing DATABASE_URL: %q: %s", conf, err)
		}
		db_name = parsed.Path[1:]
	}
	log.Printf("\tdb_path: %s\n\tdb_name: %s", db_path, db_name)

	//build our config
	config := worker.Config{
		Debug:  env("DEBUG", "") != "",
		App:    need_env("APPNAME"),
		Api:    need_env("APIKEY"),
		Name:   db_name,
		URL:    db_path,
		GOROOT: need_env("GOROOT"),
		Host:   need_env("HOST"),
	}

	//run the worker setup
	go func() {
		if err := worker.Setup(config); err != nil {
			log.Fatal("error during setup:", err)
		}
		log.Print("setup complete")
	}()

	//these handlers don't need contexts or anything special as they aren't seen by humans
	handleGet("/bins/{id}", http.HandlerFunc(handle_test_request), "test_request")
	handlePost("/bins/{id}/err", http.HandlerFunc(handle_test_error), "test_error") //more specific one has to be listed first
	handlePost("/bins/{id}", http.HandlerFunc(handle_test_response), "test_response")

	//hooks that don't need to be seen by humans
	handlePost("/hooks/github/package", http.HandlerFunc(handle_github_hook_package), "github_hook_package")
	handlePost("/hooks/github/workspace", http.HandlerFunc(handle_github_hook_workspace), "github_hook_workspace")
	handlePost("/hooks/bitbucket/package", http.HandlerFunc(handle_bitbucket_hook_package), "bitbucket_hook_package")
	handlePost("/hooks/bitbucket/workspace", http.HandlerFunc(handle_bitbucket_hook_workspace), "bitbucket_hook_workspace")
	handlePost("/hooks/google/package/{vcs}", http.HandlerFunc(handle_google_hook_package), "google_hook_package")
	handlePost("/hooks/google/workspace/{vcs}", http.HandlerFunc(handle_google_hook_workspace), "google_hook_workspace")

	//debug handler
	handleRequest("/foo", handlerFunc(handle_simple_work), "foo")

	handleGet("/build/{id}", handlerFunc(handle_build_info), "build_info")
	// handleGet("/current/html", handlerFunc(handle_work_html), "current_html")
	handleGet("/current", handlerFunc(handle_work_json), "current")
	handleGet("/recent/html", handlerFunc(handle_recent_html), "recent_html")
	handleGet("/recent", handlerFunc(handle_recent_json), "recent")
	handleGet("/how", cache(handlerFunc(handle_how)), "how")

	//this needs to go last due to how the gorilla/mux package matches (first rather than most)
	handleRequest("/", handlerFunc(handle_index), "index")

	//build the nav and subnav
	base_meta.Nav = navList{
		&navBase{"Recent", reverse("index"), nil},
		// &navBase{"Projects", reverse("index"), nil},
		&navBase{"How", reverse("how"), nil},
	}
	base_meta.SubNav = navList{}

	//set up our router
	http.Handle("/", router)
	serve_static("/assets", asset_root(""))
	if err := http.ListenAndServe(":"+env("PORT", "9080"), nil); err != nil {
		log.Fatal(err)
	}
}
