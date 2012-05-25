package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
)

func exists(cmd string) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		log.Fatal(cmd, " does not exist")
	}
}

func noexists(cmd string) {
	_, err := exec.LookPath(cmd)
	if err == nil {
		log.Fatal(cmd, " exists")
	}
}

const inst_cmd = `
	python "dist/virtualenv-1.7/virtualenv.py" --python python2.7 --distribute --never-download venv
	. venv/bin/activate
	pip install --use-mirrors mercurial
	pip install --use-mirrors bzr
`

func main() {
	//maybe have to run a shell script? :(
	//unfortunately the activate script depends on the path we run it in
	//perhaps we just call . bin/activate with the path set correctly

	os.Clearenv()
	os.Setenv("PATH", "./bin:./venv/bin")

	//check that hg fails but bash and python exist
	noexists("hg")
	noexists("bzr")

	//tools we need
	exists("bash")
	exists("python2.7")
	exists("python")
	exists("basename")

	//run bash with out shell script as stdin
	var obuf, ebuf bytes.Buffer
	bash := exec.Command("bash")
	bash.Stdin = strings.NewReader(inst_cmd)
	bash.Stdout = &obuf
	bash.Stderr = &ebuf
	err := bash.Run()

	defer os.RemoveAll("venv")

	log.Println("obuf:", obuf.String())
	log.Println("ebuf:", ebuf.String())
	if err != nil {
		log.Fatal("err:", err)
	}

	exists("hg")
	exists("bzr")

	log.Println("success!")
}
