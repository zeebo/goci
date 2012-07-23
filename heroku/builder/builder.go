package main

import (
	"code.google.com/p/gorilla/pat"
	"github.com/zeebo/goci/builder"
	bweb "github.com/zeebo/goci/builder/web"
	rweb "github.com/zeebo/goci/runner/web"
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

	//ListenAndServe!
	http.Handle("/", defaultRouter)
	bail(http.ListenAndServe(":"+env("PORT", "9080"), nil))
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

	tracker := env("TRACKER", "http://goci.me/tracker")

	//set up a builder and host it
	builder := bweb.New(
		defaultBuilder,
		tracker,
		urlWithPath("/builder"),
	)
	handleRequest("/builder", builder, "builder")
	if err := builder.Announce(); err != nil {
		bail(err)
	}
	cleanup.attach(func() {
		builder.Remove()
	})

	//set up the runner
	runner := rweb.New(
		env("APP_NAME", "goci"),
		env("API_KEY", "foo"),
		tracker,
		urlWithPath("/runner"),
	)
	handleRequest("/runner", runner, "runner")
	if err := runner.Announce(); err != nil {
		bail(err)
	}
	cleanup.attach(func() {
		runner.Remove()
	})
}
