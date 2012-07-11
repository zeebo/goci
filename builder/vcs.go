package builder

import (
	"bytes"
	"fmt"
	"github.com/zeebo/goci/environ"
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
	"bzr": vcsBzr,
}

//findVcs looks for the metadata folder inside of a given path
func findVcs(path string) (v vcs) {
	for name, vcs := range vcsMap {
		p := fp.Join(path, "."+name)
		if World.Exists(p) {
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
	Filter func(string) (bool, string)
}

var vcsGit = &vcsInfo{
	Name: "git",

	FClone:    "clone {repo} {dir}",
	FCheckout: "checkout {rev}",
	FCurrent:  "rev-parse HEAD",
	FDate:     "log -1 --format=%cD {rev}",

	Format: "Mon, 2 Jan 2006 15:04:05 -0700",
}

var vcsHg = &vcsInfo{
	Name: "hg",

	FClone:    "clone -U {repo} {dir}",
	FCheckout: "update -r {rev}",
	FCurrent:  "parents --template {node}",
	FDate:     "parents --template {date|rfc822date} -r {rev}",

	Format: time.RFC822Z,
}

var vcsBzr = &vcsInfo{
	Name: "bzr",

	FClone:    "branch {repo} {dir}",
	FCheckout: "update -r {rev}",
	FCurrent:  "revision-info --tree",
	FDate:     "log -r {rev}",

	Format: "Mon 2006-01-02 15:04:05 -0700",
	Filter: func(in string) (ok bool, out string) {
		if ok = strings.HasPrefix(in, "timestamp: "); ok {
			out = in[11:]
		}
		return
	},
}

type vcsError struct {
	Msg    string
	Vcs    *vcsInfo
	Err    error
	Args   []string
	Output string
}

func (v *vcsError) Error() string {
	return fmt.Sprintf("%s: %v\nvcs: %s\nargs: %s\noutput: %s", v.Msg, v.Err, v.Vcs.Name, v.Args, v.Output)
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

var cachedPath = map[string]string{}

func (v *vcsInfo) expandCmd(dir string, w io.Writer, cmd string, keyval ...string) (p environ.Proc, args []string) {
	vals := map[string]string{}
	for i := 0; i < len(keyval); i += 2 {
		vals[keyval[i]] = keyval[i+1]
	}

	var path string
	if ch, ok := cachedPath[v.Name]; ok {
		path = ch
	} else {
		if pa, err := World.LookPath(v.Name); err != nil {
			path = v.Name
		} else {
			path, cachedPath[v.Name] = pa, pa
		}
	}

	args = append([]string{v.Name}, expandSplit(cmd, vals)...)
	p = World.Make(environ.Command{
		W:    w,
		Dir:  dir,
		Env:  os.Environ(),
		Path: path,
		Args: args,
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

	//only get the first word
	if strings.Contains(rev, " ") {
		rev = rev[:strings.Index(rev, " ")]
	}

	return
}

func (v *vcsInfo) Date(dir, rev string) (t time.Time, err error) {
	var buf bytes.Buffer
	cmd, args := v.expandCmd(dir, &buf, v.FDate, "rev", rev)
	if e, ok := cmd.Run(); e != nil || !ok {
		err = vcsErrorf(e, v, args, buf.String(), "failed to get date")
		return
	}

	//turn the response into a number of lines
	response := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(response) == 0 {
		err = vcsErrorf(nil, v, args, buf.String(), "empty response")
		return
	}

	//set up our loop
	var i int
	line := response[i]
	ok := v.Filter == nil //don't bother looping if we have no filter

	//loop over the lines until one filters positive
	for i := 0; i < len(response) && !ok; i++ {
		ok, line = v.Filter(response[i])
	}

	//check if a line was sucessfully filtered
	if !ok {
		err = vcsErrorf(nil, v, args, buf.String(), "filter did not match")
		return
	}

	//parse the filtered line
	t, e := time.Parse(v.Format, line)
	if e != nil {
		err = vcsErrorf(e, v, args, buf.String(), "failed to parse date")
		return
	}

	return
}
