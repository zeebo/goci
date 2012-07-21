package main

import (
	"fmt"
	"net/http"
)

//handleGet sets up the handler to match a GET request of the pattern with the 
//givnen name.
func handleGet(pattern string, handler http.Handler, name string) {
	defaultRouter.Add("GET", pattern, handler).Name(name)
}

//handlePost sets up the handler to match a POST request of the pattern with the
//given name.
func handlePost(pattern string, handler http.Handler, name string) {
	defaultRouter.Add("POST", pattern, handler).Name(name)
}

//handleRequest sets the handler to match a GET or POST request of the pattern
//with the given name.
func handleRequest(pattern string, handler http.Handler, name string) {
	handleGet(pattern, handler, name)
	handlePost(pattern, handler, name)
}

//reverse turns a name and set of parameters into a URL
func reverse(name string, things ...interface{}) string {
	//convert the parameters to strings
	strs := make([]string, len(things))
	for i, th := range things {
		strs[i] = fmt.Sprint(th)
	}

	//look up the route and panic if theres a problem
	u, err := defaultRouter.GetRoute(name).URL(strs...)
	if err != nil {
		bail(err)
	}
	return u.Path
}
