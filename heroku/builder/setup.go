package main

import (
	"github.com/zeebo/goci/builder"
	"github.com/zeebo/goci/environ"
	hsetup "github.com/zeebo/goci/heroku/setup"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type LocalWorld interface {
	TempDir(string) (string, error)
	LookPath(string) (string, error)
}

var World LocalWorld = environ.New()

func valid() (err error) {
	//make sure we can find "go", "hg", and "bzr"
	for _, cmd := range []string{"go", "hg", "bzr"} {
		if _, err = World.LookPath(cmd); err != nil {
			return
		}
	}
	return
}

func setup() (err error) {
	//if everything is valid, we're done. no setup to do except the builder
	if err = valid(); err == nil {
		//find the go command again
		var path string
		path, err = World.LookPath("go")
		if err != nil {
			//something crazy happened because we just found it
			bail(err)
		}

		//get our goroot two directories up
		goroot := filepath.Dir(filepath.Dir(path))
		defaultBuilder = builder.New("", "", goroot)
		return
	}

	//make a temporary directory for our setup
	dir, err := World.TempDir("setup")
	if err != nil {
		return
	}

	//change our setup to use the same world as us, and reset it when we're done
	defer func(e hsetup.LocalWorld) { hsetup.World = e }(hsetup.World)
	hsetup.World = World.(hsetup.LocalWorld)

	//concurrently install Go and hg+bzr
	errs := make(chan error, 2)
	var gobin, vcbin string

	//launch the goroutine for installing go
	go func() {
		var err error
		gobin, err = hsetup.InstallGo(runtime.GOOS, runtime.GOARCH, dir)
		errs <- err
	}()

	//launch the goroutine for installing the vcs
	go func() {
		var err error
		vcbin, err = hsetup.InstallVCS(dir)
		errs <- err
	}()

	//return the first error we get
	for i := 0; i < cap(errs); i++ {
		err = <-errs
		if err != nil {
			return
		}
	}

	//add the bin directories to our path
	path := []string{os.Getenv("PATH"), gobin, vcbin}
	os.Setenv("PATH", strings.Join(path, string(os.PathListSeparator)))

	//make sure we're valid now
	if err = valid(); err != nil {
		return
	}

	//create our defaultBuilder with the popped GOBIN as the GOROOT
	defaultBuilder = builder.New("", "", filepath.Dir(gobin))

	//we're all set up!
	return
}
