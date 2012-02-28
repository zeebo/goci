package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	cmd := exec.Command("git", "clone", string(r), r.Dir())
	return cmd.Run()
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

func (r Repo) goCommand(args ...string) (cmd *exec.Cmd, err error) {
	root, err := Root(goVersion)
	if err != nil {
		return
	}
	cmdPath := filepath.Join(root, "bin", "go")
	cmd = exec.Command(cmdPath, args...)
	cmd.Dir = r.Dir()
	cmd.Env = []string{
		fmt.Sprintf("GOPATH=%s", cmd.Dir),
		fmt.Sprintf("GOROOT=%s", root),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}
	return
}

func (r Repo) Packages() (p []string, err error) {
	cmd, err := r.goCommand("list", "./...")
	if err != nil {
		return
	}
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	logger.Println("list:", cmd)
	err = cmd.Run()
	if err != nil {
		return
	}

	for {
		s, err := stdout.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		s = strings.TrimSpace(s)
		if len(s) > 0 {
			p = append(p, s)
		}
	}

	return
}

//Get runs go get on the package to install it and it's dependencies
func (r Repo) Get() (stdout, stderr bytes.Buffer, err error) {
	cmd, err := r.goCommand("get", "-v", "./...")
	if err != nil {
		return
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	logger.Println("get:", cmd)
	err = cmd.Run()
	return
}

//Test run's the go test command and returns the output generated and any errors
func (r Repo) Test(packages []string) (stdout, stderr bytes.Buffer, err error) {
	command := append([]string{"test", "-v"}, packages...)
	cmd, err := r.goCommand(command...)
	if err != nil {
		return
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	logger.Println("test:", cmd)
	err = cmd.Run()
	return
}
