package test

import (
	"fmt"
	"github.com/zeebo/goci/app/httputil"
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/tracker"
	"github.com/zeebo/goci/app/workqueue"
	"net/http"
)

func init() {
	http.Handle("/_test/lease", httputil.Handler(lease))
	http.Handle("/_test/ping", httputil.Handler(ping))
	http.Handle("/_test/addwork", httputil.Handler(addwork))
}

func lease(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	b, r, err := tracker.LeasePair(ctx)
	if err != nil {
		e = httputil.Errorf(err, "error leasing pair")
		return
	}

	fmt.Fprintf(w, "%+v\n%+v\n", b, r)
	return
}

func ping(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	if err := tracker.DefaultTracker.Ping(req, nil, nil); err != nil {
		e = httputil.Errorf(err, "error sending ping")
		return
	}
	fmt.Fprintf(w, "ping!")
	return
}

func addwork(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	//create our little work item
	q := rpc.Work{
		Revision:    "7d401890e4bea430b7d54a5ecd8ea50f2aea6eb2",
		ImportPath:  "github.com/zeebo/goci",
		Subpackages: true,
	}

	//add it to the queue
	if err := workqueue.QueueWork(ctx, q); err != nil {
		e = httputil.Errorf(err, "error adding work item to queue")
		return
	}

	return
}
