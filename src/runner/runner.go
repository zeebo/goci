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
	"strings"
	"time"
)

func init() {
	log.SetFlags(0)
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

type env struct {
	bin_url, src_url, post_url, err_url string
}

func newEnv(base string) env {
	return env{
		bin_url:  base,
		src_url:  base + "/src",
		post_url: base,
		err_url:  base + "/err",
	}
}

func (e env) post_error(err error, msg string) {
	error_message := fmt.Sprintf("%s: %v", msg, err)
	log.Println(error_message)
	http.Post(e.err_url, "text/plain", strings.NewReader(error_message))
}

func (e env) download(in, fn string) (name string, n int64) {
	bin, err := ioutil.TempFile("", fn)
	if err != nil {
		e.post_error(err, "error creating temp file")
		return
	}
	name = bin.Name()

	//set the file as executable
	err = os.Chmod(name, 0777)
	if err != nil {
		e.post_error(err, "error changing permissions on "+fn)
		return
	}

	resp, err := http.Get(in)
	if err != nil {
		e.post_error(err, "error downloading "+fn)
		return
	}
	defer resp.Body.Close()

	n, err = io.Copy(bin, resp.Body)
	if err != nil {
		e.post_error(err, "error copying response body into "+fn)
		return
	}

	err = bin.Sync()
	if err != nil {
		e.post_error(err, "error flushing to disk the "+fn)
		return
	}

	err = bin.Close()
	if err != nil {
		e.post_error(err, "error closing the "+fn)
		return
	}

	return
}

func main() {
	if len(os.Args) != 2 {
		log.Print("usage: runner <url to binary>")
		return
	}
	e := newEnv(os.Args[1])
	dir, err := os.Getwd()
	if err != nil {
		e.post_error(err, "unable to get current working directory")
		return
	}

	bin_name, n := e.download(e.bin_url, "test")
	defer os.Remove(bin_name)
	if n == 0 {
		e.post_error(nil, "no downloaded data")
		return
	}

	tarball, n := e.download(e.src_url, "src")
	defer os.Remove(tarball)
	if n != 0 {
		//we need to make a tmpdir with the src to run the test in
		dir, err = ioutil.TempDir("", "testdir")
		if err != nil {
			e.post_error(err, "unable to create temp dir for the src")
			return
		}
		defer os.RemoveAll(dir)

		//extract the tarball to the source directory
		cmd := exec.Command("tar", "zxf", tarball, "-C", dir)
		if err := cmd.Run(); err != nil {
			e.post_error(err, "unable to untar source")
			return
		}
	}

	var buf bytes.Buffer
	cmd := exec.Command(bin_name, "-test.v")
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Dir = dir
	cmd.Env = []string{
		//copy in some basic env vars if we have them
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("GOROOT=%s", os.Getenv("GOROOT")),
		fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH")),
	}

	//only allow the test to run for one minute
	finished := timeout(cmd, time.Minute)

	if finished {
		http.Post(e.post_url, "text/plain", &buf)
	} else {
		e.post_error(timeout_error, "error running command")
	}
}
