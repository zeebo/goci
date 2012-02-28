package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
)

var goLock = new(sync.Mutex)

type lambda func()

func with(m *sync.Mutex) lambda {
	m.Lock()
	return func() {
		m.Unlock()
	}
}

//Root returns the appropriate GOROOT path for the specified version, downloading
//the go binary release for that version if necessary.
func Root(version string) (goroot string, err error) {
	defer with(goLock)()

	//we're looking for cacheDir/go.%s.linux-amd64/go
	fullVersion := fmt.Sprintf("go.%s.%s", version, goHost)

	//set up the goroot
	goroot = filepath.Join(cacheDir, fullVersion)

	//check for the go command
	_, err = exec.LookPath(filepath.Join(goroot, "bin", "go"))
	if err != nil {

		//extract the tarball
		path := fmt.Sprintf("http://go.googlecode.com/files/%s.tar.gz", fullVersion)
		err = Extract(path, goroot)
	}

	//should be done
	return
}
