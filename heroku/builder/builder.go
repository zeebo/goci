package builder

import (
	"code.google.com/p/gorilla/pat"
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

func init() {
	//the rpcServer speaks jsonrpc
	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
}

//defaultBuilder is the builder we use and created by the setup function.
var defaultBuilder builder.Builder

//defaultRouter is the router we use to host and reverse urls.
var defaultRouter = pat.New()

//bail is a helper function to run cleanup and panic
func bail(v interface{}) {
	cleanup.cleanup()
	if v == nil {
		os.Exit(0)
	}
	panic(v)
}

func main() {
	//async run the setup and when that finishes announce and start loops
	go performInit()

	//set up the signal handler to bail and run cleanup
	go handleSignals()

	//add our handlers
	handlePost("/rpc", rpcServer, "rpc")
	handleGet("/download/{id}", http.HandlerFunc(download), "download")

	//ListenAndServe!
	http.Handle("/", defaultRouter)
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

//handleSignals sets up a channel listening on some signals and will bail when
//receiving any of them.
func handleSignals() {
	signals := []os.Signal{
		syscall.SIGQUIT,
		syscall.SIGKILL,
		syscall.SIGINT,
	}
	ch := make(chan os.Signal, len(signals))
	signal.Notify(ch, signals...)
	sig := <-ch
	log.Printf("Captured a %v\n", sig)
	bail(nil)
}

//performInit performs the application initialization: set up the builder and
//announce our capabilities to the tracker, starting the runloops if it was 
//sucessful.
func performInit() {
	if err := setup(); err != nil {
		bail(err)
	}
	if err := announce(); err != nil {
		bail(err)
	}

	//run the builder+runner loops after a sucessful announce
	go buildLoop()
	go runLoop()
}
