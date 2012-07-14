package main

import (
	"code.google.com/p/gorilla/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"net/http"
)

var tracker = client.New(
	env("TRACKER", "http://goci.me/tracker"),
	http.DefaultClient,
	client.JsonCodec,
)

func announce() {

}
