package builder

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var GOROOT = os.Getenv("GOROOT")

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
	cmd := gopathCmd(gopath, "test", "-c", pack)
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

func get(gopath string, packs ...string) (err error) {
	cmd := gopathCmd(gopath, "get", "-v", packs...)
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

func list(gopath string) (packs []string, err error) {
	cmd := gopathCmd(gopath, "list", "./...")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Dir = gopath
	err = cmd.Run()
	if err != nil {
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
	err = os.RemoveAll(dst)
	if err != nil {
		return
	}

	err = os.MkdirAll(dst, 0777)
	if err != nil {
		return
	}

	cmd := exec.Command("cp", "-a", src, dst)
	err = cmd.Run()
	if err != nil {
		err = &cmdError{
			Msg:    "Error copying files",
			Err:    err.Error(),
			Args:   cmd.Args,
			Output: "",
		}
	}
	return
}
