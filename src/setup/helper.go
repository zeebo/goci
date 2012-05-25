package setup

import "os/exec"

func exists(cmd string) (ex bool) {
	_, err := exec.LookPath(cmd)
	ex = err == nil
	return
}
