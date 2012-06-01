package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func handle_error(msg string, err error, url string) {
	if err == nil {
		return
	}

	error_message := fmt.Sprintf("%s: %s", msg, err.Error())
	http.Post(url, "text/plain", strings.NewReader(error_message))
	os.Exit(1)
}

func main() {
	if len(os.Args) != 4 {
		log.Print("usage: runner <url to binary> <url to post response> <url to post error>")
		return
	}
	bin_url, post, error_url := os.Args[1], os.Args[2], os.Args[3]

	bin, err := ioutil.TempFile("", "test")
	handle_error("error creating temp file", err, error_url)

	defer os.Remove(bin.Name())

	//set the file as executable
	err = os.Chmod(bin.Name(), 0777)
	handle_error("error changing permissions on binary", err, error_url)

	resp, err := http.Get(bin_url)
	handle_error("error downloading binary", err, error_url)

	defer resp.Body.Close()

	_, err = io.Copy(bin, resp.Body)
	handle_error("error copying response body into binary", err, error_url)

	err = bin.Sync()
	handle_error("error flusing binary to disk", err, error_url)

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

	//TODO: do this with a timeout
	cmd.Run()
	http.Post(post, "text/plain", &buf)
}
