package test

import (
	"appengine"
	"fmt"
	"httputil"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"tracker"
)

func init() {
	http.Handle("/_test/add/", http.StripPrefix("/_test/add/", httputil.Handler(add)))
	http.Handle("/_test/lease/", http.StripPrefix("/_test/lease/", httputil.Handler(lease)))
	http.Handle("/_test/lease_all", httputil.Handler(lease_all))
}

func parse(path string, args ...*string) {
	parts := strings.Split(path, "/")
	min := len(parts)
	if len(args) < min {
		min = len(args)
	}
	for i := 0; i < min; i++ {
		*args[i] = parts[i]
	}
	return
}

func add(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	var Type, GOOS, GOARCH string
	parse(req.URL.Path, &Type, &GOOS, &GOARCH)
	if GOOS == "" {
		GOOS = runtime.GOOS
	}
	if GOARCH == "" {
		GOARCH = runtime.GOARCH
	}

	//set up the request
	args := &tracker.AnnounceArgs{
		GOOS:   GOOS,
		GOARCH: GOARCH,
		Type:   Type,
		URL:    fmt.Sprintf("foo%d", rand.Int()),
	}
	var reply tracker.AnnounceReply
	ctx.Infof("%+v", args)

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

func lease(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	var GOOS, GOARCH string
	parse(req.URL.Path, &GOOS, &GOARCH)

	bk, rk, b, r, err := tracker.LeasePair(ctx, GOOS, GOARCH)
	if err != nil {
		e = httputil.Errorf(err, "error leasing pair")
		return
	}

	fmt.Fprintf(w, "%+v\n%+v\n%v\n%v", b, r, bk, rk)
	return
}

func lease_all(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	do := func(GOOS, GOARCH string) (er *httputil.Error) {
		bk, rk, b, r, err := tracker.LeasePair(ctx, GOOS, GOARCH)
		if err != nil {
			er = httputil.Errorf(err, "error leasing pair")
			return
		}
		fmt.Fprintf(w, "%+v\n%+v\n%v\n%v\n\n", b, r, bk, rk)
		return
	}

	if e = do("", ""); e != nil {
		return
	}
	if e = do("darwin", ""); e != nil {
		return
	}
	if e = do("", "amd64"); e != nil {
		return
	}
	if e = do("darwin", "amd64"); e != nil {
		return
	}

	return
}
