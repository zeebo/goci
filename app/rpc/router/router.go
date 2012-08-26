//package router provides registering and looking up rpc services
package router

import (
	"code.google.com/p/gorilla/rpc"
	"code.google.com/p/gorilla/rpc/json"
	"net/http"
	"sort"
)

//servers is a map from rpc type names to urls
var servers = map[string]string{}

//Serve adds the rpc service t to the net/http DefaultServeMux at the given path
//and allows future lookup with name.
func Serve(t interface{}, name string, path string) {
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(t, name)
	http.Handle(path, s)
	servers[name] = path
}

//Lookup returns where the rpc service lives in the net/http DefaultServeMux.
func Lookup(name string) string {
	return servers[name]
}

//Services returns the names of the registered rpc services.
func Services() []string {
	ret := make([]string, 0, len(servers))
	for name := range servers {
		ret = append(ret, name)
	}
	sort.Strings(ret)
	return ret
}
