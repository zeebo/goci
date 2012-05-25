package setup

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	fp "path/filepath"
	"runtime"
	"sync"
)

var toolLock sync.Mutex

func EnsureTool(goroot string) (err error) {
	toolLock.Lock()
	defer toolLock.Unlock()

	//fast path: tool exists
	if toolExists(goroot) {
		return
	}

	//gotta download/unzip it
	err = download(goroot)
	return
}

func toolExists(goroot string) (ex bool) {
	ex = exists("go")
	return
}

func download(goroot string) (err error) {
	//download and extract the go tarball into /usr/local/go
	tmpDir, err := ioutil.TempDir("", "go-tool")
	if err != nil {
		err = fmt.Errorf("error creating temp dir to download tool: %s", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	tbPath := fp.Join(tmpDir, "go.tar.gz")
	tarball, err := os.Create(tbPath)
	if err != nil {
		err = fmt.Errorf("error creating tmpdir: %s", err)
		return
	}
	//tarball will be cleaned by os.RemoveAll

	//create the request for the tarball
	tmpl := "https://go.googlecode.com/files/go1.0.1.%s-%s.tar.gz"
	resp, err := http.Get(fmt.Sprintf(tmpl, runtime.GOOS, runtime.GOARCH))
	if err != nil {
		err = fmt.Errorf("error downloading file: %s", err)
		return
	}
	defer resp.Body.Close()

	//write from the http response into the tarball
	_, err = io.Copy(tarball, resp.Body)
	if err != nil {
		err = fmt.Errorf("error copying to destination: %s", err)
		return
	}

	//extract the tarball
	cmd := exec.Command("tar", "zxvf", tbPath, goroot)
	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("error untarring file: %s", err)
	}
	return
}
