package setup

import (
	"os"
	"os/exec"
)

func exists(cmd string) (ex bool) {
	_, err := exec.LookPath(cmd)
	ex = err == nil
	return
}

func env(key, def string) (res string) {
	if res = os.Getenv(key); res == "" {
		res = def
	}
	return
}
