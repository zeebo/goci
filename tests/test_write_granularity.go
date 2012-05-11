package main

import (
	"log"
	"os/exec"
)

type LogWriter struct {
	what string
}

func (l LogWriter) Write(p []byte) (n int, err error) {
	log.Printf("Write[%s]: %q", l.what, p)
	n = len(p)
	return
}

func main() {
	cmd := exec.Command("go", "test", "-v", "builder")
	cmd.Stdout = LogWriter{"stdout"}
	cmd.Stderr = LogWriter{"stderr"}
	cmd.Run()
	log.Println("Done")
}
