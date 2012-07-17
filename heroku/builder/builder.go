package main

import (
	"code.google.com/p/gorilla/rpc"
	"code.google.com/p/gorilla/rpc/json"
	"net/http"
	"os"
)

//env gets an environment variable with a default
func env(key, def string) (r string) {
	if r = os.Getenv(key); r == "" {
		r = def
	}
	return
}

//rpcServer is the rpc server for interacting with the builder
var rpcServer = rpc.NewServer()

func init() {
	//the rpcServer speaks jsonrpc
	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
}

//bail is a helper function to run cleanup and panic
func bail(v interface{}) {
	cleanup.cleanup()
	if v != nil {
		panic(v)
	} else {
		os.Exit(0)
	}
}

func main() {
	//async run the setup and when that finishes announce
	go func() {
		if err := setup(); err != nil {
			bail(err)
		}
		announce()
	}()
	http.Handle("/rpc", rpcServer)
	bail(http.ListenAndServe(":9080", nil))
}

//builder is a simple goroutine that 
func builder() {
	for {
		//get a task from the queue
		task := queue.pop()
		process(task)
	}
}
