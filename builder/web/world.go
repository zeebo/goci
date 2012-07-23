package web

import (
	"github.com/zeebo/goci/environ"
	"io"
)

type LocalWorld interface {
	Open(string) (io.ReadCloser, error)
}

var World LocalWorld = environ.New()
