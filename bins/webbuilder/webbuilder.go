package main

import (
	"github.com/zeebo/goci/builder"
	"github.com/zeebo/goci/builder/web"
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

func main() {
	bu := web.New(
		builder.New(env("GOOS", ""), env("GOARCH", ""), ""),
		env("TRACKER", "http://goci.me/rpc/tracker"),
		env("HOSTED", "http://worker.goci.me/builder/"),
	)
	go http.ListenAndServe(":"+env("PORT", "9080"), bu)
	if err := bu.Announce(); err != nil {
		panic(err)
	}
	defer bu.Remove()

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
