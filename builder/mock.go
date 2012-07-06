package builder

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

var world environ = defaultEnviron{}

type environ interface {
	Exists(string) bool
	LookPath(string) (string, error)
	TempDir(string) (string, error)
}

type defaultEnviron struct{}

func (defaultEnviron) Exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func (defaultEnviron) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func (defaultEnviron) TempDir(prefix string) (string, error) {
	return ioutil.TempDir("", prefix)
}

type command struct {
	w    io.Writer
	dir  string
	env  []string
	path string
	args []string
}

type maker func(command) proc

//makeCommand makes a proc and is a variable so it can be stubbed out by the tests
var makeCommand maker = func(c command) (p proc) {
	cmd := &exec.Cmd{
		Path: c.path,
		Args: c.args,
		Env:  c.env,
	}
	if c.dir != "" {
		cmd.Dir = c.dir
	}
	if c.w != nil {
		cmd.Stdout, cmd.Stderr = c.w, c.w
	}
	return procCmd{cmd}
}

type proc interface {
	Run() (error, bool)
}

type procCmd struct {
	*exec.Cmd
}

func (p procCmd) Run() (err error, success bool) {
	err = p.Cmd.Run()
	success = p.ProcessState.Success()
	return
}
