package main

import (
	"code.google.com/p/gorilla/rpc"
	"code.google.com/p/gorilla/rpc/json"
	"net/http"
	"os"
)

func env(key, def string) (r string) {
	if r = os.Getenv(key); r == "" {
		r = def
	}
	return
}

var rpcServer = rpc.NewServer()

func init() {
	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
}

func bail(v interface{}) {
	defer cleanup.cleanup()
	panic(v)
}

func main() {
	go func() {
		if err := setup(); err != nil {
			bail(err)
		}
		announce()
	}()
	defer cleanup.cleanup()

	http.Handle("/rpc", rpcServer)
	bail(http.ListenAndServe(":9080", nil))
}
