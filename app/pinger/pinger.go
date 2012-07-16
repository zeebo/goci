package pinger

import (
	"github.com/zeebo/goci/app/rpc"
	"net/http"
)

//Pinger is a service that just has a Ping method to see if something exists.
type Pinger struct{}

//Ping is a simple method to see if the rpc exists.
func (Pinger) Ping(req *http.Request, args *rpc.None, rep *rpc.None) (err error) {
	return
}
