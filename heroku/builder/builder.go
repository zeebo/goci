package main

import (
	"code.google.com/p/gorilla/rpc"
	"code.google.com/p/gorilla/rpc/json"
	"github.com/zeebo/goci/builder"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

//defaultBuilder is the builder we use and created by the setup function.
var defaultBuilder builder.Builder

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
		if err := announce(); err != nil {
			bail(err)
		}

		//run the builder+runner loops after a sucessful announce
		go buildLoop()
		go runLoop()
	}()

	//set up the signal handler to bail and run cleanup
	signals := []os.Signal{
		syscall.SIGQUIT,
		syscall.SIGKILL,
		syscall.SIGINT,
	}
	ch := make(chan os.Signal, len(signals))
	signal.Notify(ch, signals...)
	go func() {
		sig := <-ch
		log.Printf("Captured a %v\n", sig)
		bail(nil)
	}()

	http.Handle("/rpc", rpcServer)
	bail(http.ListenAndServe(":9080", nil))
}

//buildLoop is a simple goroutine that grabs items from the queue and sends them
//off for processing.
func buildLoop() {
	for {
		//get a task from the queue
		task := builderQueue.Pop()
		process(task)
	}
}

//runLoop is a simple goroutine that grabs items from the queue and sends them
//off for processing.
func runLoop() {
	for {
		task := runnerQueue.Pop()
		process_run(task)
	}
}
