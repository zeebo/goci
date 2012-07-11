//setup provides functionality to build the Go distribution from source
package setup

import (
	"errors"
	"fmt"
	"github.com/zeebo/goci/environ"
	"github.com/zeebo/goci/tarball"
	"net/http"
	"os"
	"runtime"
	fp "path/filepath"
)

type LocalWorld interface {
	MkdirAll(string, os.FileMode) error
	Make(environ.Command) environ.Proc
}

var World LocalWorld = environ.New()
var ErrInvalidGOROOT = errors.New("GOROOT must end with go")

func InstallGo(dir string) (err error) {
	dir = fp.Clean(dir)
	if fp.Base(dir) != "go" {
		err = ErrInvalidGOROOT
		return
	}

	//drop of the go part
	ext := fp.Dir(dir)

	vers := runtime.Version()
	src := fmt.Sprintf("http://go.googlecode.com/files/%s.src.tar.gz", vers)

	resp, err := http.Get(src)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	//make the goroot
	if err = World.MkdirAll(dir, 0666); err != nil {
		return
	}

	err = tarball.Extract(resp.Body, ext)
	if err != nil {
		return
	}

	return
}
