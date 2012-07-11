package environ

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

//New returns a new environment.
func New() (e Env) { return }

//Env is the default environment that provides wrappers for os and system
//functions with side effects.
type Env struct{}

//Exists attempts to os.Stat the name and returns if there was an error.
func (Env) Exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

//LookPath is a wrapper for exec.LookPath
func (Env) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

//TempDir is a wrapper for ioutil.TempDir
func (Env) TempDir(prefix string) (string, error) {
	return ioutil.TempDir("", prefix)
}

//Stat is a wrapper for os.Lstat.
func (Env) Stat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

//Readdir Opens the path provided and reads all the infos for the directory.
func (Env) Readdir(path string) (s []os.FileInfo, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	return f.Readdir(-1)
}

//Create is a wrapper for os.Create but using the specified os.FileMode
func (Env) Create(path string, mode os.FileMode) (io.WriteCloser, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
}

//Open is a wrapper for os.Open
func (Env) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

//MkdirAll is a wrapper for os.MkdirAll
func (Env) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

//Make creates a Proc from a Command that allows you to execute processes.
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

//Command is a type that represents the information for executing a command.
type Command struct {
	W    io.Writer
	Dir  string
	Env  []string
	Path string
	Args []string
}

//Proc is the type returned by the environments Make method that lets you run
//the command.
type Proc interface {
	Run() (error, bool)
}

//procCmd is the internal representation of a process.
type procCmd struct {
	*exec.Cmd
}

//Run makes procCmd ad Proc and just hands it off to the wrapped exec command.
func (p procCmd) Run() (err error, success bool) {
	err = p.Cmd.Run()
	success = err == nil && p.ProcessState.Success()
	return
}
