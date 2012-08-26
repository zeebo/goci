package main

import (
	//rpc handlers
	"github.com/zeebo/goci/app/response"   //handle responses from workers
	"github.com/zeebo/goci/app/rpc/router" //to register the rpc handlers
	"github.com/zeebo/goci/app/tracker"    //handle tracking workers

	//debug handlers
	"expvar"                           //for some exported variables
	_ "github.com/zeebo/goci/app/test" //simple test handlers
	_ "net/http/pprof"                 //add pprof support

	//normal handlers
	"github.com/zeebo/goci/app/frontend"    //load up the web frontend for people
	_ "github.com/zeebo/goci/app/workqueue" //handle queuing/dispatching work
	"net/http"
)

//services returns the names of the services and satisfies the expvar.Func type
func services() interface{} { return router.Services() }

//add our exported vars
func init() {
	expvar.Publish("Services", expvar.Func(services))
}

//add our rpc services
func init() {
	router.Serve(response.Response{}, "Response", "/rpc/response")
	router.Serve(tracker.Tracker{}, "Tracker", "/rpc/tracker")
}

//add the frontend to the DefaultServeMux. Should be called after environment
//has been loaded.
func addFrontend() (err error) {
	//tell it where the templates live and compile them
	frontend.Config.Templates = env("TEMPLATES", "")
	if err = frontend.Compile(); err != nil {
		return err
	}

	//add it in
	http.Handle("/", frontend.Handler)
}
