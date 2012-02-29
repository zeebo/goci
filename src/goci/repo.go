package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var extGOPATH = filepath.Join(cacheDir, "gopath")

func init() {
	if err := os.MkdirAll(extGOPATH, 0777); err != nil {
		panic(err)
	}
}

//Repo is a path to a repository like "git://github.com/zeebo/goci.git"
type Repo string

//Name returns the name of the repo with the last 4 chars (.git) removed
func (r Repo) Name() string {
	name := path.Base(string(r))
	return name[:len(name)-4]
}

//Hash returns the hex represntation of the sha1 sum of the repo.
func (r Repo) Hash() string {
	hash := sha1.New()
	hash.Write([]byte(r))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

//Clone clones the repo into the temporary directory
func (r Repo) Clone() (err error) {
	cmd := exec.Command("git", "clone", string(r), r.Dir())
	err = cmd.Run()
	if err != nil {
		return
	}

	//check for a src directory
	_, err = os.Stat(filepath.Join(r.Dir(), "src"))
	if err == nil {
		return
	}

	logger.Println(r, "non GOPATH repo dectected. Moving to GOPATH")

	//we had an error so let's reset that and move things
	//around so that we have it in a gopath
	repoDir := filepath.Join(cacheDir, r.Name())
	cmd = exec.Command("mv", r.Dir(), repoDir)
	if err = cmd.Run(); err != nil {
		return
	}

	//make the directory
	srcDir := filepath.Join(r.Dir(), "src")
	if err = os.MkdirAll(srcDir, 0777); err != nil {
		return
	}

	//move it into src, and we're done
	cmd = exec.Command("mv", repoDir, srcDir+string(filepath.Separator))
	err = cmd.Run()
	return
}

//Cleanup removes the checked out repository
func (r Repo) Cleanup() error {
	return exec.Command("rm", "-rf", r.Dir()).Run()
}

//Checkout checks out the repository to a specific commit
func (r Repo) Checkout(commit string) error {
	cmd := exec.Command("git", "checkout", commit)
	//this requires some thought
	cmd.Dir = r.GitDir()
	return cmd.Run()
}

//Dir returns the directory of where the repo will be cloned.
func (r Repo) Dir() string {
	return filepath.Join(cacheDir, r.Hash())
}

func (r Repo) GitDir() string {
	//check for a .git directory in r.Dir()
	_, err := os.Stat(filepath.Join(r.Dir(), ".git"))
	if err == nil {
		return r.Dir()
	}
	return filepath.Join(r.Dir(), "src", r.Name())
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
		fmt.Sprintf("GOPATH=%s:%s", extGOPATH, cmd.Dir),
		fmt.Sprintf("GOROOT=%s", root),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}
	logger.Println("running:", cmd.Dir, cmd.Args, cmd.Env)
	return
}

func (r Repo) Packages() (p []string, err error) {
	cmd, err := r.goCommand("list", "./...")
	if err != nil {
		return
	}
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
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

//Clean runs the go clean command on all the packages passed in
func (r Repo) Clean(packages []string) (stdout, stderr bytes.Buffer, err error) {
	command := append([]string{"clean", "-i"}, packages...)
	cmd, err := r.goCommand(command...)
	if err != nil {
		return
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	return
}

//Get runs go get on the package to install it and it's dependencies
func (r Repo) Get(packages []string) (stdout, stderr bytes.Buffer, err error) {
	command := append([]string{"get", "-v"}, packages...)
	cmd, err := r.goCommand(command...)
	if err != nil {
		return
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
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
	err = cmd.Run()
	return
}

//TestInstall runs a test -i on the packages
func (r Repo) TestInstall(packages []string) {
	command := append([]string{"test", "-i"}, packages...)
	cmd, err := r.goCommand(command...)
	if err != nil {
		return
	}
	cmd.Run()
}
