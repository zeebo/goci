package builder

import (
	"bytes"
	"fmt"
	"github.com/zeebo/goci/environ"
	"io"
	"strings"
	fp "path/filepath"
	p "path"
	tb "github.com/zeebo/goci/tarball"
)

type cmdError struct {
	Msg    string
	Err    error
	Args   []string
	Output string
}

func (t *cmdError) Error() string {
	return fmt.Sprintf("%s: %v\nargs: %s\noutput: %s", t.Msg, t.Err, t.Args, t.Output)
}

func cmdErrorf(err error, args []string, buf string, format string, vals ...interface{}) error {
	return &cmdError{
		Msg:    fmt.Sprintf(format, vals...),
		Err:    err,
		Args:   args,
		Output: buf,
	}
}

const listTemplate = `{{ range .TestImports }}{{ . }}
{{ end }}{{ range .XTestImports }}{{ . }}
{{ end }}`

func tarball(dir, out string) (err error) {
	err = tb.CompressFile(dir, out)
	return
}

//goCmd uses the builders goroot variable to get a path to the go command with
//the correct environment
func (b Builder) goCmd(buf io.Writer, dir string, args ...string) (p environ.Proc) {
	cmd := environ.Command{
		W:    buf,
		Dir:  dir,
		Env:  b.env,
		Path: fp.Join(b.goroot, "bin", "go"),
		Args: args,
	}
	return world.Make(cmd)
}

//goRun is a convenience wrapper that builds and executes a command, returning
//an error that wraps all the output.
func (b Builder) goRun(dir string, msg string, args ...string) (s string, err error) {
	var buf bytes.Buffer
	if e, ok := b.goCmd(&buf, dir, args...).Run(); !ok {
		err = cmdErrorf(e, args, buf.String(), "error %s", msg)
	}
	s = buf.String()
	return
}

//goGet runs a go get on the given import paths
func (b Builder) goGet(download bool, path ...string) (err error) {
	//build the arguments
	args := []string{"go", "get", "-v"}
	if download {
		args = append(args, "-d")
	}
	args = append(args, "-tags", "goci")
	args = append(args, path...)

	_, err = b.goRun("", "building code + deps", args...)
	return
}

//goList runs a list on the given import paths and returns the paths that match
//the list query as well as all
func (b Builder) goList(path string) (paths, testpaths []string, err error) {
	s, err := b.goRun(b.gopath, "listing package", "go", "list", path)
	if err != nil {
		return
	}
	paths = parseImports(s)

	s, err = b.goRun(b.gopath, "listing package", "go", "list", "-f", listTemplate, path)
	if err != nil {
		return
	}
	testpaths = parseImports(s)

	return
}

func (b Builder) goTest(path string) (bin string, err error) {
	dir, err := world.TempDir("build")
	if err != nil {
		return
	}

	_, err = b.goRun(dir, "building test", "go", "test", "-c", "-tags", "goci", path)
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
