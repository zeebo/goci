package builder

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os/exec"
	"strings"
)

type VCS interface {
	Checkout(dir, rev string) (err error)
	Clone(repo, dir string) (err error)
}

func init() {
	gob.Register(&vcs{})
}

type vcs struct {
	Name string

	FClone    string
	FCheckout string
}

var (
	Git VCS = vcsGit //Git is a VCS for git based repositories
	HG  VCS = vcsHg  //Hg is a VCS for mercurial based repositories
)

var (
	vcsGit = &vcs{
		Name: "git",

		FClone:    "clone {repo} {dir}",
		FCheckout: "checkout {rev}",
	}

	vcsHg = &vcs{
		Name: "hg",

		FClone:    "clone -U {repo} {dir}",
		FCheckout: "update -r {rev}",
	}
)

type vcsError struct {
	msg    string
	vcs    *vcs
	err    error
	args   []string
	output string
}

func (v *vcsError) Error() string {
	return fmt.Sprintf("%s: %s\nvcs: %s\nargs: %s\noutput: %s", v.msg, v.err.Error(), v.vcs.Name, v.args, v.output)
}

func expand(s string, vals map[string]string) string {
	for k, v := range vals {
		s = strings.Replace(s, "{"+k+"}", v, -1)
	}
	return s
}

func expandSplit(s string, vals map[string]string) []string {
	return strings.Split(expand(s, vals), " ")
}

func (v *vcs) expandCmd(cmd string, keyval ...string) (c *exec.Cmd) {
	vals := map[string]string{}
	for i := 0; i < len(keyval); i += 2 {
		vals[keyval[i]] = keyval[i+1]
	}
	c = exec.Command(v.Name, expandSplit(cmd, vals)...)
	return
}

func (v *vcs) Checkout(dir, rev string) (err error) {
	cmd := v.expandCmd(v.FCheckout, "rev", rev)
	var buf bytes.Buffer
	cmd.Dir = dir
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err = cmd.Run()
	if err != nil {
		err = &vcsError{
			msg:    "Failed to checkout",
			vcs:    v,
			err:    err,
			args:   cmd.Args,
			output: buf.String(),
		}
	}
	return
}

func (v *vcs) Clone(repo, dir string) (err error) {
	cmd := v.expandCmd(v.FClone, "repo", repo, "dir", dir)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err = cmd.Run()
	if err != nil {
		err = &vcsError{
			msg:    "Failed to clone",
			vcs:    v,
			err:    err,
			args:   cmd.Args,
			output: buf.String(),
		}
	}
	return
}
