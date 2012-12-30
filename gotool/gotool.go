package gotool

import (
	"bytes"
	"fmt"
	"github.com/zeebo/goci/environ"
	"io"
	p "path"
	fp "path/filepath"
	"strings"
)

type CmdError struct {
	Msg    string
	Err    error
	Args   []string
	Output string
}

func (t *CmdError) Error() string {
	return fmt.Sprintf("%s: %v\nargs: %s\noutput: %s", t.Msg, t.Err, t.Args, t.Output)
}

func cmdErrorf(err error, args []string, buf string, format string, vals ...interface{}) error {
	return &CmdError{
		Msg:    fmt.Sprintf(format, vals...),
		Err:    err,
		Args:   args,
		Output: buf,
	}
}

const listTemplate = `{{ range .TestImports }}{{ . }}
{{ end }}{{ range .XTestImports }}{{ . }}
{{ end }}`

type Gotool struct {
	Env    []string
	GOROOT string
	GOPATH string
}

//Cmd uses the goroot variable to get a path to the go command with the correct
//environment
func (g *Gotool) Cmd(buf io.Writer, dir string, args ...string) (p environ.Proc) {
	cmd := environ.Command{
		W:    buf,
		Dir:  dir,
		Env:  g.Env,
		Path: fp.Join(g.GOROOT, "bin", "go"),
		Args: args,
	}
	return World.Make(cmd)
}

//Run is a convenience wrapper that builds and executes a command, returning
//an error that wraps all the output.
func (g *Gotool) Run(dir string, msg string, args ...string) (s string, err error) {
	var buf bytes.Buffer
	if e, ok := g.Cmd(&buf, dir, args...).Run(); !ok {
		err = cmdErrorf(e, args, buf.String(), "error %s", msg)
	}
	s = buf.String()
	return
}

//Get runs a go get on the given import paths
func (g *Gotool) Get(download bool, path ...string) (err error) {
	//build the arguments
	args := []string{"go", "get", "-v"}
	if download {
		args = append(args, "-d")
	}
	args = append(args, "-tags", "goci")
	args = append(args, path...)

	_, err = g.Run("", "building code + deps", args...)
	return
}

//List runs a list on the given import paths and returns the paths that match
//the list query as well as all
func (g *Gotool) List(path string) (paths, testpaths []string, err error) {
	s, err := g.Run(g.GOPATH, "listing package", "go", "list", path)
	if err != nil {
		return
	}
	paths = parseImports(s)

	s, err = g.Run(g.GOPATH, "listing package", "go", "list", "-f", listTemplate, path)
	if err != nil {
		return
	}
	testpaths = parseImports(s)

	return
}

//Test creates a test binary for the import path. exeSuffix should be ".exe"
//if running on windows, and "" otherwise.
func (g *Gotool) Test(exeSuffix, path string) (bin string, err error) {
	dir, err := World.TempDir("build")
	if err != nil {
		return
	}

	_, err = g.Run(dir, "building test", "go", "test", "-c", "-tags", "goci", path)
	if err != nil {
		return
	}

	//what the go tool does from inspecting the source
	_, elem := p.Split(path)
	bin = fp.Join(dir, elem+".test"+exeSuffix)
	return
}

func parseImports(data string) (imps []string) {
	for _, p := range strings.Split(data, "\n") {
		if strings.HasPrefix(p, "_") {
			continue
		}
		if tr := strings.TrimSpace(p); len(tr) > 0 {
			imps = append(imps, tr)
		}
	}
	imps = unique(imps)
	return
}
