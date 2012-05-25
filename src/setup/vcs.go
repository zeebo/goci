package setup

import "sync"

var vcsLock sync.Mutex

func EnsureVCS() (err error) {
	vcsLock.Lock()
	defer vcsLock.Unlock()

	//fast path: if tools exist, dont do anything
	if vcsExists() {
		return
	}

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
