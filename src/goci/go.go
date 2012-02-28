package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
)

/*
   url=http://go.googlecode.com/files/$pkg_path.tar.gz
   curl -sO $url
   tar zxf $pkg_path.tar.gz
   rm -f $pkg_path.tar.gz
*/

var goLock sync.Mutex

type lambda func()

func with(m sync.Mutex) lambda {
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

	goroot = filepath.Join(cacheDir, fullVersion)
	_, err = exec.LookPath(filepath.Join(goroot, "bin", "go"))
	if err != nil {
		err = download(fullVersion)
	}
	return
}

func download(fullVersion string) (err error) {
	file := fmt.Sprintf("%s.tar.gz", fullVersion)

	cmd := exec.Command("curl", "-sO", fmt.Sprintf("http://go.googlecode.com/files/%s", file))
	cmd.Dir = cacheDir

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", cmd.Args, err)
	}

	cmd = exec.Command("tar", "zxf", file)
	cmd.Dir = cacheDir

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", cmd.Args, err)
	}

	cmd = exec.Command("mv", "go", fullVersion)
	cmd.Dir = cacheDir

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", cmd.Args, err)
	}

	cmd = exec.Command("rm", "-f", file)
	cmd.Dir = cacheDir

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", cmd.Args, err)
	}

	return
}
