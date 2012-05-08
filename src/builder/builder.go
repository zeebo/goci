package builder

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	fp "path/filepath"
	"time"
)

//Work represents an item of work to be completed by the builder
type Work interface {
	Revisions() (rev []string)
	VCS() (v VCS)
	RepoPath() (path string)
	ImportPath() (path string)
	IsWorkspace() (ok bool)
}

type Report struct {
	When     time.Time
	Duration time.Duration
	Revision string
	Passed   bool
	Output   string
	Error    error
}

func gopathCmd(gopath string, bin string, args ...string) (cmd *exec.Cmd) {
	cmd = exec.Command(bin, args...)
	cmd.Dir = gopath
	cmd.Env = []string{
		fmt.Sprintf("GOPATH=%s", gopath),
	}
	return
}

func test(gopath, pack string) (output string, ok bool, err error) {
	cmd := gopathCmd(gopath, "go", "test", "-v", pack)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	output = buf.String()
	ok = cmd.ProcessState.Success()
	return
}

func get(gopath, pack string) (err error) {
	cmd := gopathCmd(gopath, "go", "get", "-v", pack)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	e := cmd.Run()
	if !cmd.ProcessState.Success() {
		err = fmt.Errorf("Error building the code + deps: %s\noutput: %s", e, buf)
	}
	return
}

func Run(w Work) (res []Report, err error) {
	vcs := w.VCS()

	dir, err := ioutil.TempDir("", "gopath")
	if err != nil {
		return
	}
	defer os.RemoveAll(dir)

	//create gopath rooted at dir
	pack := w.ImportPath()
	srcDir := fp.Join(dir, "src", pack)
	err = vcs.Clone(w.RepoPath(), srcDir)
	if err != nil {
		return
	}

	for _, rev := range w.Revisions() {
		rep := Report{
			When:     time.Now(),
			Revision: rev,
		}

		//checkout
		rep.Error = vcs.Checkout(srcDir, rev)
		if rep.Error != nil {
			//return an error report for this revision
			rep.Duration = time.Since(rep.When)
			res = append(res, rep)
			continue
		}

		//run a get
		rep.Error = get(dir, pack)
		if rep.Error != nil {
			//return an error report for this revision
			rep.Duration = time.Since(rep.When)
			res = append(res, rep)
			continue
		}

		//run the tests
		rep.Output, rep.Passed, rep.Error = test(dir, pack)
		rep.Duration = time.Since(rep.When)
		res = append(res, rep)
	}

	return
}
