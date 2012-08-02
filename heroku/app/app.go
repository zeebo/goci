package main

import (
	"github.com/zeebo/goci/builder"
	buweb "github.com/zeebo/goci/builder/web"
	"github.com/zeebo/goci/heroku/setup"
	ruweb "github.com/zeebo/goci/runner/web"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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
	ru := ruweb.New(
		env("APP_NAME", "goci"),
		env("API_KEY", "foo"),
		env("TRACKER", "http://goci.me/tracker"),
		env("RUNHOSTED", "http://runner.goci.me/runner"),
	)
	http.Handle("/runner", ru)

	//start the server
	go http.ListenAndServe(":"+env("PORT", "9080"), nil)

	//announce the runner
	if err := ru.Announce(); err != nil {
		panic(err)
	}
	defer ru.Remove()

	//run the painful setup in parallel
	ech, bins := make(chan error), make(chan string)
	var goroot string

	//the goroot
	go func() {
		dir, err := ioutil.TempDir("", "goroot")
		if err != nil {
			ech <- err
			return
		}
		//set the goroot
		goroot = filepath.Join(dir, "go")

		//make the tempdir writable
		if err := os.Chmod(dir, 0777); err != nil {
			ech <- err
			return
		}

		bin, err := setup.InstallGo(dir)
		if err != nil {
			ech <- err
			return
		}

		ech <- nil
		bins <- bin
	}()

	//the tools
	go func() {
		dir, err := ioutil.TempDir("", "tools")
		if err != nil {
			ech <- err
			return
		}

		bin, err := setup.InstallVCS("heroku/dist", dir)
		if err != nil {
			ech <- err
			return
		}

		ech <- nil
		bins <- bin
	}()

	//grab two errors (should be nil)
	for i := 0; i < 2; i++ {
		if err := <-ech; err != nil {
			panic(err)
		}
	}

	//grab both the bin dirs (we always get 2 because we got 2 nil errors)
	for i := 0; i < 2; i++ {
		path := os.Getenv("PATH")
		bin := <-bins
		os.Setenv("PATH", path+string(filepath.ListSeparator)+bin)
	}

	//create the builder and announce it
	bu := buweb.New(
		builder.New("linux", "amd64", goroot),
		env("TRACKER", "http://goci.me/tracker"),
		env("BUILDHOSTED", "http://runner.goci.me/builder"),
	)
	http.Handle("/builder", bu)

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
