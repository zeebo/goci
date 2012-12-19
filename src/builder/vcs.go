package builder

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type VCS interface {
	Checkout(dir, rev string) (err error)
	Clone(repo, dir string) (err error)
	Current(dir string) (rev string, err error)
	Date(dir, rev string) (t time.Time, err error)
}

func init() {
	gob.Register(&vcs{})
	gob.Register(&vcsError{})
	gob.Register(&time.ParseError{})
}

type vcs struct {
	Name string

	FClone    string
	FCheckout string
	FCurrent  string
	FDate     string

	Format string
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
		FCurrent:  "rev-parse HEAD",
		FDate:     "log -1 --format=%cD {rev}",

		Format: "Mon, 2 Jan 2006 15:04:05 -0700",
	}

	vcsHg = &vcs{
		Name: "hg",

		FClone:    "clone -U {repo} {dir}",
		FCheckout: "update -r {rev}",
		FCurrent:  "parents --template {node}",
		FDate:     "parents --template {date|rfc822date} -r {rev}",

		Format: time.RFC1123Z,
	}
)

type vcsError struct {
	Msg    string
	Vcs    *vcs
	Err    error
	Args   []string
	Output string
}

func (v *vcsError) Error() string {
	return fmt.Sprintf("%s: %s\nvcs: %s\nargs: %s\noutput: %s", v.Msg, v.Err.Error(), v.Vcs.Name, v.Args, v.Output)
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

	if e := cmd.Run(); e != nil {
		err = &vcsError{
			Msg:    "Failed to checkout",
			Vcs:    v,
			Err:    e,
			Args:   cmd.Args,
			Output: buf.String(),
		}
	}
	return
}

func (v *vcs) Clone(repo, dir string) (err error) {
	cmd := v.expandCmd(v.FClone, "repo", repo, "dir", dir)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if e := cmd.Run(); e != nil {
		err = &vcsError{
			Msg:    "Failed to clone",
			Vcs:    v,
			Err:    e,
			Args:   cmd.Args,
			Output: buf.String(),
		}
	}
	return
}

func (v *vcs) Current(dir string) (rev string, err error) {
	cmd := v.expandCmd(v.FCurrent)
	var buf bytes.Buffer
	cmd.Dir = dir
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if e := cmd.Run(); e != nil {
		err = &vcsError{
			Msg:    "Failed to get current revision",
			Vcs:    v,
			Err:    e,
			Args:   cmd.Args,
			Output: buf.String(),
		}
		return
	}

	//do parsing into rev
	rev = strings.TrimSpace(buf.String())
	return
}

func (v *vcs) Date(dir, rev string) (t time.Time, err error) {
	cmd := v.expandCmd(v.FDate, "rev", rev)
	var buf bytes.Buffer
	cmd.Dir = dir
	cmd.Stdout = &buf
	cmd.Stdin = &buf

	if e := cmd.Run(); e != nil {
		err = &vcsError{
			Msg:    "Failed to get the date for revision",
			Vcs:    v,
			Err:    e,
			Args:   cmd.Args,
			Output: buf.String(),
		}
		return
	}

	//parse the time
	t, e := time.Parse(v.Format, strings.TrimSpace(buf.String()))
	if e != nil {
		err = &vcsError{
			Msg:    "Failed to parse date for revision",
			Vcs:    v,
			Err:    e,
			Args:   cmd.Args,
			Output: buf.String(),
		}
	}
	return
}
