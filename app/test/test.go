// +build !goci

package test

import (
	"appengine"
	"fmt"
	"httputil"
	"net/http"
	"queue"
	"rpc"
	"tracker"
)

func init() {
	http.Handle("/_test/lease", httputil.Handler(lease))
	http.Handle("/_test/ping", httputil.Handler(ping))
	http.Handle("/_test/addwork", httputil.Handler(addwork))
}

func lease(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	bk, rk, b, r, err := tracker.LeasePair(ctx)
	if err != nil {
		e = httputil.Errorf(err, "error leasing pair")
		return
	}

	fmt.Fprintf(w, "%+v\n%+v\n%v\n%v", b, r, bk, rk)
	return
}

func ping(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	if err := tracker.DefaultTracker.Ping(req, nil, nil); err != nil {
		e = httputil.Errorf(err, "error sending ping")
		return
	}
	fmt.Fprintf(w, "ping!")
	return
}

func addwork(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	//create our little work item
	q := rpc.Work{
		Revision:   "e9dd26552f10d390b5f9f59c6a9cfdc30ed1431c",
		ImportPath: "github.com/zeebo/irc",
	}

	q = rpc.Work{
		Revision:    "7e283bf6dbf4c97a00647f873faa0b513ad59fbf",
		ImportPath:  "github.com/zeebo/goci",
		Subpackages: true,
	}

	//add it to the queue
	if err := queue.QueueWork(ctx, q); err != nil {
		e = httputil.Errorf(err, "error adding work item to queue")
		return
	}

	return
}
