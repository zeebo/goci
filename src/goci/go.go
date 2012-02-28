package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

/*
   url=http://go.googlecode.com/files/$pkg_path.tar.gz
   curl -sO $url
   tar zxf $pkg_path.tar.gz
   rm -f $pkg_path.tar.gz
*/

func Root(version string) (goroot string, err error) {

}

func download(version string) (err error) {
	tmp := os.TempDir()
	file := fmt.Sprintf("%s.tar.gz", version)

	cmd := exec.Command("curl", "-s0", fmt.Sprintf("http://go.googlecode.com/files/%s", file))
	cmd.Dir = tmp

	if err = cmd.Run(); err != nil {
		return
	}

	cmd = exec.Command("tar", "zxf", file)
	cmd.Dir = tmp

	if err = cmd.Run(); err != nil {
		return
	}

	cmd = exec.Command("rm", "-f", file)
	cmd.Dir = tmp

	return cmd.Run()
}
