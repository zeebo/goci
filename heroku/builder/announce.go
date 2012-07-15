package main

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

func announceWithArgs(args *rpc.AnnounceArgs) {
	reply := new(rpc.AnnounceReply)
	if err := tracker.Call("Tracker.Announce", args, reply); err != nil {
		bail(err)
	}
	cleanup.attach(func() {
		args := &rpc.RemoveArgs{
			Key: reply.Key,
		}
		reply := &rpc.None{}
		tracker.Call("Tracker.Remove", args, reply)
	})
}

func announce() {
	url := env("RPC_URL", "http://builder.goci.me/rpc")

	announceWithArgs(&rpc.AnnounceArgs{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
		Type:   "Builder",
		URL:    url,
	})

	announceWithArgs(&rpc.AnnounceArgs{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
		Type:   "Runner",
		URL:    url,
	})
}
