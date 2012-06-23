package builder

import (
	"os"
	"os/exec"
)

type Bundle struct {
	Path    string
	Tarball string
}

// tar -cvzf /path_name_of_tarball/bakup.2009.4.28.tgz -C /path_name_of_dir .
func tarball(dir, out string) (err error) {
	cmd := exec.Command("tar", "-cvzf", out, "-C", dir, ".")
	err = cmd.Run()
	if err != nil {
		return
	}
	//make sure the tarball exists
	_, err = os.Stat(out)
	return
}
