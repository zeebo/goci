package main

import (
	"github.com/zeebo/goci/runner/web"
	"log"
	"net/http"
)

//env gets an environment variable with a default
func env(key, def string) (r string) {
	if r = os.Getenv(key); r == "" {
		r = def
	}
	return
}

func main() {
	runner := web.New(
		env("APP_NAME", "goci"),
		env("API_KEY", "foo"),
		env("TRACKER", "http://goci.me/tracker"),
		env("HOSTED", "http://runner.goci.me"),
	)
	go http.ListenAndServe(":"+env("PORT", "9080"), runner)
	if err := runner.Announce(); err != nil {
		panic(err)
	}
	defer runner.Remove()

	//wait for a signal
	signals := []os.Signal{
		syscall.SIGQUIT,
		syscall.SIGKILL,
		syscall.SIGINT,
	}
	ch := make(chan os.Signal, len(signals))
	signal.Notify(ch, signals...)
	sig := <-ch
	log.Printf("Captured a %v\n", sig)
}
