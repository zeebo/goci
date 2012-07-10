package pinger

import "net/http"

//Pinger is a service that just has a Ping method to see if something exists.
type Pinger struct{}

//Ping is the argument type.
type Ping struct{}

//Pong is the reply type.
type Pong struct{}

//Ping is a simple method to see if the rpc exists.
func (Pinger) Ping(req *http.Request, args *Ping, rep *Pong) (err error) {
	return
}
