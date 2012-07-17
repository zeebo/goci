package pinger

import "net/http"

//Pinger is a service that just has a Ping method to see if something exists.
type Pinger struct{}

//None is an rpc.None inlined so we dont have the dependency.
type None struct{}

//Ping is a simple method to see if the rpc exists.
func (Pinger) Ping(req *http.Request, args *None, rep *None) (err error) {
	return
}
