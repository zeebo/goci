package gotool

import (
	"github.com/zeebo/goci/environ"
)

type LocalWorld interface {
	TempDir(string) (string, error)
	Make(environ.Command) environ.Proc
}

//World allows tests and package users to stub out the environment
var World LocalWorld = environ.New()
