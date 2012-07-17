package tarball

import (
	"compress/gzip"
	"github.com/zeebo/goci/environ"
	"io"
	"os"
)

type LocalWorld interface {
	//get info about files/directories
	Stat(string) (os.FileInfo, error)
	Readdir(string) ([]os.FileInfo, error)

	//create/open files
	Create(string, os.FileMode) (io.WriteCloser, error)
	Open(string) (io.ReadCloser, error)
	MkdirAll(string, os.FileMode) error
}

var World LocalWorld = environ.New()

var compression = gzip.BestCompression
