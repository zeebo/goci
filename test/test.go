package test

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"httputil"
	"net/http"
	"runtime"
	"tracker"
)

func init() {
	http.Handle("/_test/add", httputil.Handler(add))
	http.Handle("/_test/lease/builder", httputil.Handler(lease_builder))
	http.Handle("/_test/lease/runner", httputil.Handler(lease_runner))
	http.Handle("/_test/lease/queries", httputil.Handler(lease_tons))
}

func add(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	//set up the request
	args := &tracker.AnnounceArgs{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
		Type:   "Builder",
		URL:    "foo",
	}
	var reply tracker.AnnounceReply

	//perform the call
	if err := tracker.DefaultTracker.Announce(req, args, &reply); err != nil {
		e = httputil.Errorf(err, "error calling announce")
		return
	}
	//log the reply
	ctx.Infof("%+v", reply)
	fmt.Fprintf(w, "%+v\n", reply)
	return
}

func lease(w http.ResponseWriter, req *http.Request, ctx appengine.Context, ltype string) (e *httputil.Error) {
	key, err := tracker.Lease(ctx, "", "", ltype)
	if err == tracker.ErrNoneAvailable {
		fmt.Fprintln(w, "no service of that type available")
		return
	}
	if err != nil {
		e = httputil.Errorf(err, "unable to lease %s", ltype)
		return
	}
	s := new(tracker.Service)
	if err := datastore.Get(ctx, key, s); err != nil {
		e = httputil.Errorf(err, "unable to load leased client")
		return
	}
	fmt.Fprintf(w, "%+v", s)
	return
}

func lease_builder(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	return lease(w, req, ctx, "Builder")
}

func lease_runner(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	return lease(w, req, ctx, "Runner")
}

func lease_tons(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	tracker.Lease(ctx, "", "", "a")
	tracker.Lease(ctx, "f", "", "a")
	tracker.Lease(ctx, "", "f", "a")
	tracker.Lease(ctx, "f", "f", "a")
	fmt.Fprintln(w, "ran 4 queries. all lease indicies should be built")
	return
}
