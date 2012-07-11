package environ

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func New() (e Env) { return }

type Env struct{}

func (Env) Exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func (Env) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func (Env) TempDir(prefix string) (string, error) {
	return ioutil.TempDir("", prefix)
}

func (Env) Stat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

func (Env) Readdir(path string) (s []os.FileInfo, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	return f.Readdir(-1)
}

func (Env) Create(path string, mode os.FileMode) (io.WriteCloser, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
}

func (Env) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (Env) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func (Env) Make(c Command) (p Proc) {
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
	return procCmd{cmd}
}

type Command struct {
	W    io.Writer
	Dir  string
	Env  []string
	Path string
	Args []string
}

type Proc interface {
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
