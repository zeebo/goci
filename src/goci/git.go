package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

//Repo is a path to a repository like "http://github.com/zeebo/goci"
type Repo string

//Hash returns the hex represntation of the sha1 sum of the repo.
func (r Repo) Hash() string {
	hash := sha1.New()
	hash.Write([]byte(r))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

//Clone clones the repo into the temporary directory
func (r Repo) Clone() (err error) {
	return exec.Command("git", "clone", string(r), r.Dir()).Run()
}

//Dir returns the directory of where the repo will be cloned.
func (r Repo) Dir() string {
	//silencing the error and just using a relative path if theres ever an
	//issue getting the current working directory
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "tmp", r.Hash())
}

//Test run's the go test command and returns the output generated and any errors
func (r Repo) Test() (out bytes.Buffer, err error) {
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Dir = r.Dir()
	cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%s", cmd.Dir))
	cmd.Stdout = &out
	err = cmd.Run()
	return
}
