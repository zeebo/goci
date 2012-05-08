package builder

import (
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

func expand(s string, vals map[string]string) string {
	for k, v := range vals {
		s = strings.Replace(s, "{"+k+"}", v, -1)
	}
	return s
}

func expandSplit(s string, vals map[string]string) []string {
	return strings.Split(expand(s, vals), " ")
}

func (v *vcs) expandCmd(cmd string, keyval ...string) *exec.Cmd {
	vals := map[string]string{}
	for i := 0; i < len(keyval); i += 2 {
		vals[keyval[i]] = keyval[i+1]
	}
	return exec.Command(v.name, expandSplit(cmd, vals)...)
}

func (v *vcs) Checkout(dir, rev string) (err error) {
	cmd := v.expandCmd(v.checkout, "rev", rev)
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		err = format(cmd, err)
	}
	return
}

func (v *vcs) Clone(repo, dir string) (err error) {
	cmd := v.expandCmd(v.clone, "repo", repo, "dir", dir)
	err = cmd.Run()
	if err != nil {
		err = format(cmd, err)
	}
	return
}

func format(cmd *exec.Cmd, err error) (e error) {
	e = fmt.Errorf("Error running %s: %s", strings.Join(cmd.Args, " "), err)
	return
}
