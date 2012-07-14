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

func announce_with_args(args *rpc.AnnounceArgs) {
	reply := new(rpc.AnnounceReply)
	if err := tracker.Call("Tracker.Announce", args, reply); err != nil {
		panic(err)
	}
	cleanup.attach(func() {
		args := &rpc.RemoveArgs{
			Key: reply.Key,
		}
		reply := &rpc.RemoveReply{}
		tracker.Call("Tracker.Remove", args, reply)
	})
}

func announce() {
	announce_with_args(&rpc.AnnounceArgs{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
		Type:   "Builder",
		URL:    env("RPC_URL", "http://builder.goci.me/rpc"),
	})

	announce_with_args(&rpc.AnnounceArgs{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
		Type:   "Runner",
		URL:    env("RPC_URL", "http://builder.goci.me/rpc"),
	})
}
