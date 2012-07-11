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

func (Env) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

func (Env) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (Env) MkdirAll(path string) error {
	return os.MkdirAll(path, 0777)
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
	return ProcCmd{cmd}
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

type ProcCmd struct {
	*exec.Cmd
}

func (p ProcCmd) Run() (err error, success bool) {
	err = p.Cmd.Run()
	success = p.ProcessState.Success()
	return
}
