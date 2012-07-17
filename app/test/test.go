// +build !goci

package test

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"httputil"
	"net/http"
	"queue"
	"rpc"
	"time"
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
	q := &queue.Work{
		Work: rpc.Work{
			Revision:   "foo",
			ImportPath: "bar",
		},
		Data:    "foo",
		Created: time.Now(),
	}

	key := datastore.NewIncompleteKey(ctx, "Work", nil)
	var err error
	if key, err = datastore.Put(ctx, key, q); err != nil {
		e = httputil.Errorf(err, "error storing work item")
		return
	}

	if err = queue.AddQueue(ctx, httputil.ToString(key)); err != nil {
		e = httputil.Errorf(err, "error adding work item to queue")
		return
	}

	return
}
