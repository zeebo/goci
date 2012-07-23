package builder

import (
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"net/http"
	"runtime"
)

var tracker = client.New(
	env("TRACKER", "http://goci.me/tracker"),
	http.DefaultClient,
	client.JsonCodec,
)

func announceWithArgs(args *rpc.AnnounceArgs) (err error) {
	reply := new(rpc.AnnounceReply)
	if err = tracker.Call("Tracker.Announce", args, reply); err != nil {
		return
	}
	cleanup.attach(func() {
		args := &rpc.RemoveArgs{
			Key: reply.Key,
		}
		reply := &rpc.None{}
		tracker.Call("Tracker.Remove", args, reply)
	})
	return
}

func announce() (err error) {
	url := urlWithPath(reverse("rpc"))

	err = announceWithArgs(&rpc.AnnounceArgs{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
		Type:   "Builder",
		URL:    url,
	})
	if err != nil {
		return
	}

	err = announceWithArgs(&rpc.AnnounceArgs{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
		Type:   "Runner",
		URL:    url,
	})
	if err != nil {
		return
	}

	return
}
