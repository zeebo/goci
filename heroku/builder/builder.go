package main

import (
	"code.google.com/p/gorilla/rpc"
	"code.google.com/p/gorilla/rpc/json"
	"log"
	"net/http"
	"os"
)

func env(key, def string) (r string) {
	if r = os.Getenv(key); r == "" {
		r = def
	}
	return
}

var rpc_server = rpc.NewServer()

func init() {
	rpc_server.RegisterCodec(json.NewCodec(), "application/json")
}

func main() {
	go func() {
		if err := setup(); err != nil {
			log.Panic(err)
		}
		announce()
	}()
	defer cleanup.cleanup()

	http.Handle("/rpc", rpc_server)
	panic(http.ListenAndServe(":9080", nil))
}
