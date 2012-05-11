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

func vcsExists() (ex bool) {
	cmd := exec.Command("hg", "--version")
	if ex = cmd.Run() == nil; !ex {
		return
	}
	cmd = exec.Command("bzr", "version")
	ex = cmd.Run() == nil
	return
}
