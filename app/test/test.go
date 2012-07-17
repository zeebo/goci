// +build !goci

package test

import (
	"appengine"
	"fmt"
	"httputil"
	"math/rand"
	"net/http"
	"rpc"
	"runtime"
	"strings"
	"tracker"
)

func init() {
	http.Handle("/_test/lease", httputil.Handler(lease))
	http.Handle("/_test/ping", httputil.Handler(ping))
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
