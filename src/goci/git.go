package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

//Repo is a path to a repository like "git://github.com/zeebo/goci.git"
type Repo string

//Hash returns the hex represntation of the sha1 sum of the repo.
func (r Repo) Hash() string {
	hash := sha1.New()
	hash.Write([]byte(r))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

//Clone clones the repo into the temporary directory
func (r Repo) Clone() error {
	return exec.Command("git", "clone", string(r), r.Dir()).Run()
}

//Cleanup removes the checked out repository
func (r Repo) Cleanup() error {
	return exec.Command("rm", "-rf", r.Dir()).Run()
}

//Checkout checks out the repository to a specific commit
func (r Repo) Checkout(commit string) error {
	cmd := exec.Command("git", "checkout", commit)
	cmd.Dir = r.Dir()
	return cmd.Run()
}

//Dir returns the directory of where the repo will be cloned.
func (r Repo) Dir() string {
	return filepath.Join(os.TempDir(), r.Hash())
}

//Get runs go get on the package to install it and it's dependencies
func (r Repo) Get() (stdout, stderr bytes.Buffer, err error) {
	root, err := Root(goVersion)
	if err != nil {
		return
	}
	cmdPath := filepath.Join(root, "bin", "go")
	cmd := exec.Command(cmdPath, "get", "-v", "all")
	cmd.Dir = r.Dir()
	cmd.Env = []string{
		fmt.Sprintf("GOPATH=%s", cmd.Dir),
		fmt.Sprintf("GOROOT=%s", root),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	return
}

//Test run's the go test command and returns the output generated and any errors
func (r Repo) Test() (stdout, stderr bytes.Buffer, err error) {
	root, err := Root(goVersion)
	if err != nil {
		return
	}
	cmdPath := filepath.Join(root, "bin", "go")
	cmd := exec.Command(cmdPath, "test", "-v", "./...")
	cmd.Dir = r.Dir()
	cmd.Env = []string{
		fmt.Sprintf("GOPATH=%s", cmd.Dir),
		fmt.Sprintf("GOROOT=%s", root),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	return
}
