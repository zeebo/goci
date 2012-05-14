package setup

import (
	"os/exec"
	"sync"
)

var vcsLock sync.Mutex

func EnsureVCS() (err error) {
	vcsLock.Lock()
	defer vcsLock.Unlock()

	//fast path: if hg + bzr exist, dont do anything
	if vcsExists() {
		return
	}

	return
}

func vcsExists() bool {
	tools := []string{"hg", "bzr", "git"}
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			return false
		}
	}
	return true
}
