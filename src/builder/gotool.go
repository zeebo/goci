package builder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	GOROOT = env("GOROOT", "/usr/local/go")
	PATH   = env("PATH", "/usr/bin:/usr/local/bin:/usr/local/go/bin")
)

func gopathCmd(gopath, action, arg string, args ...string) (cmd *exec.Cmd) {
	if args == nil {
		cmd = exec.Command("go", action, arg)
	} else {
		cmd = exec.Command("go", append([]string{action, arg}, args...)...)
	}
	cmd.Dir = gopath
	cmd.Env = []string{
		fmt.Sprintf("GOPATH=%s", gopath),
		fmt.Sprintf("PATH=%s", PATH),
	}
	return
}

func test(gopath string, packs ...string) (output string, ok bool, err error) {
	cmd := gopathCmd(gopath, "test", "-v", packs...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	output = buf.String()
	ok = cmd.ProcessState.Success()
	return
}

func get(gopath string, packs ...string) (err error) {
	cmd := gopathCmd(gopath, "get", "-v", packs...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	e := cmd.Run()
	if !cmd.ProcessState.Success() {
		err = fmt.Errorf("Error building the code + deps: %s\nargs: %s\noutput: %s", e, cmd.Args, buf)
	}
	return
}

func list(gopath string) (packs []string, err error) {
	cmd := gopathCmd(gopath, "list", "./...")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	if err != nil {
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

	cmd := exec.Command("cp", "-r", src, dst)
	err = cmd.Run()
	return
}
