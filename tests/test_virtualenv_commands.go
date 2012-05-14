package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os/exec"
	fp "path/filepath"
)

var (
	dist_dir      = "/Users/zeebo/Code/pkg/goci/dist"
	virtualenv_py = fp.Join(dist_dir, "virtualenv-1.7/virtualenv.py")
	python        = "python2.7"
)

func virtualenv(args ...string) (cmd *exec.Cmd) {
	args = append([]string{virtualenv_py}, args...)
	cmd = exec.Command("python", args...)
	return
}

func install() {
	venv, err := ioutil.TempDir("", "venv")
	if err != nil {
		log.Fatal(err)
	}

	cmd := virtualenv("--python", python, "--distribute", "--never-download", "--prompt='(venv) '", venv)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		log.Fatal(err, buf.String())
	}
	log.Println(buf.String())
}

func main() {
	//maybe have to run a shell script? :(
	//unfortunately the activate script depends on the path we run it in
	//perhaps we just call . bin/activate with the path set correctly

	//check that hg fails
	pth, err := exec.LookPath("hg")
	log.Println(pth, err)
}
