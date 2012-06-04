package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

var timeout_error = errors.New("timeout")

func timeout(cmd *exec.Cmd, dur time.Duration) (ok bool) {
	done := make(chan bool, 1)
	if err := cmd.Start(); err != nil {
		log.Println("error starting command:", err)
		return
	}
	defer cmd.Process.Kill()

	//start a race
	go func() {
		cmd.Wait()
		done <- true
	}()

	go func() {
		<-time.After(dur)
		done <- false
	}()

	//see who won
	ok = <-done
	return
}

func main() {
	if len(os.Args) != 2 {
		log.Print("usage: runner <url to binary>")
		return
	}
	bin_url, post, error_url := os.Args[1], os.Args[1], path.Join(os.Args[1], "err")
	var err error

	//define a little helper that closes on the error value and error url
	post_error := func(msg string) {
		error_message := fmt.Sprintf("%s: %s", msg, err.Error())
		log.Println(error_message)
		http.Post(error_url, "text/plain", strings.NewReader(error_message))
	}

	bin, err := ioutil.TempFile("", "test")
	if err != nil {
		post_error("error creating temp file")
		return
	}

	defer os.Remove(bin.Name())

	//set the file as executable
	err = os.Chmod(bin.Name(), 0777)
	if err != nil {
		post_error("error changing permissions on binary")
		return
	}

	resp, err := http.Get(bin_url)
	if err != nil {
		post_error("error downloading binary")
		return
	}
	defer resp.Body.Close()

	_, err = io.Copy(bin, resp.Body)
	if err != nil {
		post_error("error copying response body into binary")
		return
	}

	err = bin.Sync()
	if err != nil {
		post_error("error flusing binary to disk")
		return
	}

	err = bin.Close()
	if err != nil {
		post_error("error closing the binary")
		return
	}

	var buf bytes.Buffer
	cmd := exec.Command(bin.Name(), "-test.v")
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Env = []string{
		//copy in some basic env vars if we have them
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("GOROOT=%s", os.Getenv("GOROOT")),
		fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH")),
	}

	//only allow the test to run for one minute
	finished := timeout(cmd, time.Minute)

	if finished {
		http.Post(post, "text/plain", &buf)
	} else {
		err = timeout_error
		post_error("error running command")
	}
}
