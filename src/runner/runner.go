package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) != 3 {
		log.Print("usage: runner <url to binary> <url to post response>")
		return
	}
	bin_url, post := os.Args[1], os.Args[2]

	bin, err := ioutil.TempFile("", "test")
	if err != nil {
		log.Print("error creating temp file: ", err)
		return
	}
	defer os.Remove(bin.Name())

	//set the file as executable
	err = os.Chmod(bin.Name(), 0777)
	if err != nil {
		log.Print("error changing permissions on binary: ", err)
		return
	}

	resp, err := http.Get(bin_url)
	if err != nil {
		log.Print("error downloading binary: ", err)
		return
	}
	defer resp.Body.Close()

	_, err = io.Copy(bin, resp.Body)
	if err != nil {
		log.Print("error copying response body into binary: ", err)
		return
	}

	err = bin.Sync()
	if err != nil {
		log.Print("error flusing binary to disk: ", err)
		return
	}

	var buf bytes.Buffer
	cmd := exec.Command(bin.Name(), "-test.v")
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()

	if err != nil {
		log.Print("error running test binary: ", err)
		return
	}

	_, err = http.Post(post, "text/plain", &buf)
	if err != nil {
		log.Print("error posting to response url: ", err)
		return
	}
}
