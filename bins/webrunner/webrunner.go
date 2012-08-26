package main

import (
	"github.com/zeebo/goci/runner/direct"
	"github.com/zeebo/goci/runner/web"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

//Service can handle http requests and announce and remove its presence to a
//tracker.
type Service interface {
	http.Handler
	Announce() error
	Remove() error
}

//newWebRunner returns a service for running tests on the heroku dyno mesh.
func newWebRunner() Service {
	runner := web.New(
		mustEnv("APP_NAME"),
		mustEnv("API_KEY"),
		env("TRACKER", "http://goci.me/rpc/tracker"),
		mustEnv("HOSTED"),
	)
	return runner
}

//newDirectRunner returns a service for running tests on the local machine.
func newDirectRunner() Service {
	runner := direct.New(
		mustEnv("RUNNER"),
		env("TRACKER", "http://goci.me/rpc/tracker"),
		mustEnv("HOSTED"),
	)
	return runner
}

//env gets an environment variable with a default
func env(key, def string) (r string) {
	if r = os.Getenv(key); r == "" {
		r = def
	}
	return
}

//mustEnv panics if the environment variable is not set.
func mustEnv(key string) (r string) {
	if r = env(key, ""); r == "" {
		panic("Please specify the " + key + " environment variable.")
	}
	return
}

func main() {
	//create the runner based on the DIRECT variable
	var runner Service
	if env("DIRECT", "") == "" {
		runner = newWebRunner()
	} else {
		runner = newDirectRunner()
	}

	l, err := net.Listen("tcp", "0.0.0.0:"+env("PORT", "9080"))
	if err != nil {
		panic(err)
	}
	defer l.Close()
	go http.Serve(l, runner)

	if err := runner.Announce(); err != nil {
		panic(err)
	}
	defer runner.Remove()

	//wait for a signal
	signals := []os.Signal{
		syscall.SIGQUIT,
		syscall.SIGKILL,
		syscall.SIGINT,
		syscall.SIGTERM,
	}
	ch := make(chan os.Signal, len(signals))
	signal.Notify(ch, signals...)
	sig := <-ch
	log.Printf("Captured a %v\n", sig)
}
