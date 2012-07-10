package environ

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func New() Environ {
	return defaultEnviron{}
}

type Command struct {
	W    io.Writer
	Dir  string
	Env  []string
	Path string
	Args []string
}

type Environ interface {
	Exists(string) bool
	LookPath(string) (string, error)
	TempDir(string) (string, error)
	Make(Command) Proc
}

type Proc interface {
	Run() (error, bool)
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

func (defaultEnviron) Make(c Command) (p Proc) {
	cmd := &exec.Cmd{
		Path: c.Path,
		Args: c.Args,
		Env:  c.Env,
	}
	if c.Dir != "" {
		cmd.Dir = c.Dir
	}
	if c.W != nil {
		cmd.Stdout, cmd.Stderr = c.W, c.W
	}
	return ProcCmd{cmd}
}

type ProcCmd struct {
	*exec.Cmd
}

func (p ProcCmd) Run() (err error, success bool) {
	err = p.Cmd.Run()
	success = p.ProcessState.Success()
	return
}
