package builder

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	GOROOT = os.Getenv("GOROOT")
	GOPATH = os.Getenv("GOPATH")
)

func init() {
	gob.Register(&cmdError{})
}

type cmdError struct {
	Msg    string
	Err    string
	Args   []string
	Output string
}

func (t *cmdError) Error() string {
	return fmt.Sprintf("%s: %s\nargs: %s\noutput: %s", t.Msg, t.Err, t.Args, t.Output)
}

func gopathCmd(gopath, action, arg string, args ...string) (cmd *exec.Cmd) {
	if args == nil {
		cmd = exec.Command("go", action, arg)
	} else {
		cmd = exec.Command("go", append([]string{action, arg}, args...)...)
	}
	cmd.Env = []string{
		fmt.Sprintf("GOPATH=%s", gopath),
		fmt.Sprintf("GOROOT=%s", GOROOT),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}
	return
}

func testbuild(gopath, pack, dir string) (err error) {
	cmd := gopathCmd(gopath, "test", "-c", "-tags", "goci", pack)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Dir = dir
	e := cmd.Run()
	if !cmd.ProcessState.Success() {
		err = &cmdError{
			Msg:    fmt.Sprintf("Error building a %s binary", pack),
			Err:    e.Error(),
			Args:   cmd.Args,
			Output: buf.String(),
		}
	}
	return
}

func get(gopath string, update bool, packs ...string) (err error) {
	var base []string
	if update {
		base = []string{"-u", "-tags", "goci"}
	} else {
		base = []string{"-tags", "goci"}
	}

	cmd := gopathCmd(gopath, "get", "-v", append(base, packs...)...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	e := cmd.Run()
	if !cmd.ProcessState.Success() {
		err = &cmdError{
			Msg:    "Error building the code + deps",
			Err:    e.Error(),
			Args:   cmd.Args,
			Output: buf.String(),
		}
	}
	return
}

const listTemplate = `{{ range .TestImports }}{{ . }}
{{ end }}{{ range .XTestImports }}{{ . }}
{{ end }}`

func list(gopath string) (packs, testpacks []string, err error) {
	packs, testpacks, err = listPackage(gopath, "./...")
	return
}

func listPackage(gopath string, pack string) (packs, testpacks []string, err error) {
	cmd := gopathCmd(gopath, "list", pack)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Dir = gopath
	if e := cmd.Run(); e != nil {
		err = &cmdError{
			Msg:    "Error listing the packages",
			Err:    err.Error(),
			Args:   cmd.Args,
			Output: buf.String(),
		}
		return
	}

	for _, p := range strings.Split(buf.String(), "\n") {
		if strings.HasPrefix(p, "_") {
			continue
		}
		if tr := strings.TrimSpace(p); len(tr) > 0 {
			packs = append(packs, tr)
		}
	}
	packs = unique(packs)

	//list all the imports for the test files
	cmd = gopathCmd(gopath, "list", "-f", listTemplate, pack)
	buf.Reset()
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Dir = gopath
	if e := cmd.Run(); e != nil {
		err = &cmdError{
			Msg:    "Error listing the packages",
			Err:    err.Error(),
			Args:   cmd.Args,
			Output: buf.String(),
		}
		return
	}

	for _, p := range strings.Split(buf.String(), "\n") {
		if strings.HasPrefix(p, "_") {
			continue
		}
		if tr := strings.TrimSpace(p); len(tr) > 0 {
			testpacks = append(testpacks, tr)
		}
	}
	testpacks = unique(testpacks)
	return
}

func search(packs []string, p string) bool {
	for _, op := range packs {
		if p == op {
			return true
		}
	}
	return false
}

func copy(src, dst string) (err error) {
	if e := os.RemoveAll(dst); e != nil {
		err = &cmdError{
			Msg:  "Error removing all",
			Err:  e.Error(),
			Args: []string{dst},
		}

		return
	}

	if e := os.MkdirAll(dst, 0777); e != nil {
		err = &cmdError{
			Msg:  "Error making all directories",
			Err:  e.Error(),
			Args: []string{dst},
		}
		return
	}

	cmd := exec.Command("cp", "-a", src, dst)
	if e := cmd.Run(); e != nil {
		err = &cmdError{
			Msg:  "Error copying files",
			Err:  e.Error(),
			Args: cmd.Args,
		}
		return
	}

	return
}

// tar -cvzf /path_name_of_tarball/tb.tar.gz -C /path_name_of_dir .
func tarball(dir, out string) (err error) {
	cmd := exec.Command("tar", "-cvzf", out, "-C", dir, ".")
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	if e := cmd.Run(); e != nil {
		err = &cmdError{
			Msg:    "error building tarball of source",
			Err:    e.Error(),
			Args:   cmd.Args,
			Output: buf.String(),
		}
		return
	}
	//make sure the tarball exists
	if _, e := os.Stat(out); e != nil {
		err = &cmdError{
			Msg: "couldn't find tarball after making it",
			Err: e.Error(),
		}
	}
	return
}
