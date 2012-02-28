package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func init() {
	//set up the path to work like on heroku
	if err := os.Setenv("PATH", "/usr/bin"); err != nil {
		panic(err)
	}
}

//Repo is a path to a repository like "git://github.com/zeebo/goci.git"
type Repo string

//Hash returns the hex represntation of the sha1 sum of the repo.
func (r Repo) Hash() string {
	hash := sha1.New()
	hash.Write([]byte(r))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

//Clone clones the repo into the temporary directory
func (r Repo) Clone() (err error) {
	cmd := exec.Command("git", "clone", string(r), r.Dir())
	return cmd.Run()
}

func (r Repo) Cleanup() error {
	return exec.Command("rm", "-rf", r.Dir()).Run()
}

//Dir returns the directory of where the repo will be cloned.
func (r Repo) Dir() string {
	return filepath.Join(os.TempDir(), r.Hash())
}

//Test run's the go test command and returns the output generated and any errors
func (r Repo) Test() (out bytes.Buffer, err error) {
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Dir = r.Dir()
	cmd.Env = []string{fmt.Sprintf("GOPATH=%s", cmd.Dir)}
	cmd.Stdout = &out
	log.Println(cmd)
	err = cmd.Run()
	return
}
