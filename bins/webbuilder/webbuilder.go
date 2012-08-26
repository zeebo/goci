package main

import (
	"github.com/zeebo/goci/builder"
	"github.com/zeebo/goci/builder/web"
	"log"
	"net"
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
	hosted := env("HOSTED", "")
	if hosted == "" {
		panic("don't know where the builder lives. Please set the HOSTED env var.")
	}

	bu := web.New(
		builder.New(env("GOOS", ""), env("GOARCH", ""), ""),
		env("TRACKER", "http://goci.me/rpc/tracker"),
		hosted,
	)

	l, err := net.Listen("tcp", "0.0.0.0:"+env("PORT", "9080"))
	if err != nil {
		panic(err)
	}
	defer l.Close()
	go http.Serve(l, bu)

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
