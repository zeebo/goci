package builder

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type VCS interface {
	Checkout(dir, rev string) (err error)
	Clone(repo, dir string) (err error)
}

type vcs struct {
	name string

	clone    string
	checkout string
}

var (
	Git VCS = vcsGit //Git is a VCS for git based repositories
	HG  VCS = vcsHg  //Hg is a VCS for mercurial based repositories
)

var (
	vcsGit = &vcs{
		name: "git",

		clone:    "clone {repo} {dir}",
		checkout: "checkout {rev}",
	}

	vcsHg = &vcs{
		name: "hg",

		clone:    "clone -U {repo} {dir}",
		checkout: "update -r {rev}",
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
	return fmt.Sprintf("%s: %s\nvcs: %s\nargs: %s\noutput: %s", v.msg, v.err.Error(), v.vcs.name, v.args, v.output)
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
	c = exec.Command(v.name, expandSplit(cmd, vals)...)
	return
}

func (v *vcs) Checkout(dir, rev string) (err error) {
	cmd := v.expandCmd(v.checkout, "rev", rev)
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
	cmd := v.expandCmd(v.clone, "repo", repo, "dir", dir)
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
