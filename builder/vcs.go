package builder

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	fp "path/filepath"
)

type vcs interface {
	Checkout(dir, rev string) (err error)
	Clone(repo, dir string) (err error)
	Current(dir string) (rev string, err error)
	Date(dir, rev string) (t time.Time, err error)
}

var vcsMap = map[string]vcs{
	"git": vcsGit,
	"hg":  vcsHg,
	// "bzr": nil,
}

func findVcs(path string) (v vcs) {
	for name, vcs := range vcsMap {
		p := fp.Join(path, "."+name)
		if world.Exists(p) {
			v = vcs
			return
		}
	}
	return
}

type vcsInfo struct {
	Name string

	FClone    string
	FCheckout string
	FCurrent  string
	FDate     string

	Format string
}

var (
	vcsGit = &vcsInfo{
		Name: "git",

		FClone:    "clone {repo} {dir}",
		FCheckout: "checkout {rev}",
		FCurrent:  "rev-parse HEAD",
		FDate:     "log -1 --format=%cD {rev}",

		Format: "Mon, 2 Jan 2006 15:04:05 -0700",
	}

	vcsHg = &vcsInfo{
		Name: "hg",

		FClone:    "clone -U {repo} {dir}",
		FCheckout: "update -r {rev}",
		FCurrent:  "parents --template {node}",
		FDate:     "parents --template {date|rfc822date} -r {rev}",

		Format: time.RFC822Z,
	}
)

type vcsError struct {
	Msg    string
	Vcs    *vcsInfo
	Err    error
	Args   []string
	Output string
}

func (v *vcsError) Error() string {
	return fmt.Sprintf("%s: %s\nvcs: %s\nargs: %s\noutput: %s", v.Msg, v.Err.Error(), v.Vcs.Name, v.Args, v.Output)
}

func vcsErrorf(err error, v *vcsInfo, args []string, out string, msg string) error {
	return &vcsError{
		Msg:    msg,
		Vcs:    v,
		Err:    err,
		Args:   args,
		Output: out,
	}
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

func (v *vcsInfo) expandCmd(dir string, w io.Writer, cmd string, keyval ...string) (p proc, args []string) {
	vals := map[string]string{}
	for i := 0; i < len(keyval); i += 2 {
		vals[keyval[i]] = keyval[i+1]
	}
	path, err := world.LookPath(v.Name)
	if err != nil {
		path = v.Name
	}
	args = append([]string{v.Name}, expandSplit(cmd, vals)...)
	p = world.Make(command{
		w:    w,
		dir:  dir,
		env:  os.Environ(),
		path: path,
		args: args,
	})
	return
}

func (v *vcsInfo) Checkout(dir, rev string) (err error) {
	var buf bytes.Buffer
	cmd, args := v.expandCmd(dir, &buf, v.FCheckout, "rev", rev)
	if e, ok := cmd.Run(); e != nil || !ok {
		err = vcsErrorf(e, v, args, buf.String(), "failed to checkout")
	}
	return
}

func (v *vcsInfo) Clone(repo, dir string) (err error) {
	var buf bytes.Buffer
	cmd, args := v.expandCmd("", &buf, v.FClone, "repo", repo, "dir", dir)
	if e, ok := cmd.Run(); e != nil || !ok {
		err = vcsErrorf(e, v, args, buf.String(), "failed to clone")
	}
	return
}

func (v *vcsInfo) Current(dir string) (rev string, err error) {
	var buf bytes.Buffer
	cmd, args := v.expandCmd(dir, &buf, v.FCurrent)
	if e, ok := cmd.Run(); e != nil || !ok {
		err = vcsErrorf(e, v, args, buf.String(), "failed to get revision")
		return
	}

	//do parsing into rev
	rev = strings.TrimSpace(buf.String())
	return
}

func (v *vcsInfo) Date(dir, rev string) (t time.Time, err error) {
	var buf bytes.Buffer
	cmd, args := v.expandCmd(dir, &buf, v.FDate, "rev", rev)
	if e, ok := cmd.Run(); e != nil || !ok {
		err = vcsErrorf(e, v, args, buf.String(), "failed to get date")
		return
	}

	//parse the time
	t, e := time.Parse(v.Format, strings.TrimSpace(buf.String()))
	if e != nil {
		err = vcsErrorf(e, v, args, buf.String(), "failed to parse date")
	}
	return
}
