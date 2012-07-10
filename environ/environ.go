package environ

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func New() interface{} {
	return defaultEnviron{}
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

func (defaultEnviron) Stat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

func (defaultEnviron) Readdir(path string) (s []os.FileInfo, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	return f.Readdir(-1)
}

func (defaultEnviron) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

func (defaultEnviron) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
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
