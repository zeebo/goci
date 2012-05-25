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
	//clear out our env
	os.Clearenv()
	os.Setenv("PATH", "./bin:./venv/bin")

	//make sure we dont have hg/bzr yet
	noexists("hg")
	noexists("bzr")

	//tools we need to install
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

	//be a tidy citizen! :)
	defer os.RemoveAll("venv")

	//do some logging
	log.Println("obuf:", obuf.String())
	log.Println("ebuf:", ebuf.String())
	if err != nil {
		log.Fatal("err:", err)
	}

	//make sure it worked
	exists("hg")
	exists("bzr")

	//yay!
	log.Println("success!")
}
