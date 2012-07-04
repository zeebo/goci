package test

import (
	"appengine"
	"appengine/urlfetch"
	"httputil"
	"net/http"
	"rpc/client"
	"runtime"
	"time"
	"tracker"
)

func init() {
	http.Handle("/_test/add", httputil.Handler(add))
}

func add(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	//create a client
	hcl := urlfetch.Client(ctx)
	hcl.Transport.(*urlfetch.Transport).Deadline = 1 * time.Second
	cl := client.New("http://127.1:8080/tracker", hcl, client.JsonCodec)

	//set up the request
	args := &tracker.AnnounceArgs{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
		Type:   "Builder",
		URL:    "foo",
	}
	var reply tracker.AnnounceReply

	//perform the call
	if err := cl.Call("Announce.Announce", args, &reply); err != nil {
		e = httputil.Errorf(err, "error on rpc call")
		return
	}

	//log the reply
	ctx.Infof("%+v", reply)

	return
}
