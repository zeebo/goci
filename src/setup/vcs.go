package setup

import (
	"fmt"
	"os"
	"os/exec"
	fp "path/filepath"
	"strings"
	"sync"
)

var vcsLock sync.Mutex

func EnsureVCS() (err error) {
	vcsLock.Lock()
	defer vcsLock.Unlock()

	//fast path: if tools exist, dont do anything
	if vcsExists() {
		return
	}

	//otherwise install them
	err = vcsInstall()
	return
}

func vcsExists() bool {
	tools := []string{"hg", "bzr"}
	for _, tool := range tools {
		if !exists(tool) {
			return false
		}
	}
	return true
}

func vcsInstall() (err error) {
	//make sure the dist dir exists
	if _, err = os.Stat(distdir); err != nil {
		return
	}

	//make sure we have the commands we need
	cmds := []string{"bash", "python2.7", "python", "basename"}
	for _, cmd := range cmds {
		if !exists(cmd) {
			err = fmt.Errorf("%s: command not found", cmd)
			return
		}
	}

	vcs_inst_cmd := fmt.Sprintf(`
		python "%s" --python python2.7 --distribute --never-download %s
		. %s
		pip install --use-mirrors mercurial
		pip install --use-mirrors bzr
	`,
		fp.Join(distdir, "virtualenv-1.7", "virtualenv.py"),
		venvdir,
		fp.Join(venvdir, "bin", "activate"),
	)

	bash := exec.Command("bash")
	bash.Stdin = strings.NewReader(vcs_inst_cmd)
	err = bash.Run()

	if err != nil {
		return
	}

	//success! - first see if we have the tools.
	//if we don't, add venvdir/bin to the path
	if !vcsExists() {
		path := fmt.Sprintf("%s:%s", os.Getenv("PATH"), fp.Join(venvdir, "bin"))
		os.Setenv("PATH", path)
	}
	//if we still don't we have an error
	if !vcsExists() {
		err = fmt.Errorf("script ran but can't find hg+bzr anywhere")
		return
	}
	return
}
