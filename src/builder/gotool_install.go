package builder

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

func ensureTool() (err error) {
	toolLock.Lock()
	defer toolLock.Unlock()

	//where to download the go binaries
	goroot := "/usr/local/go"

	//fast path: tool exists
	if toolExists(goroot) {
		return
	}

	//gotta download/unzip it
	err = download(goroot)
	return
}

func toolExists(goroot string) (ex bool) {
	//just try to run go version
	cmd := exec.Command("go", "version")
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s", fp.Join(goroot, "bin")),
	}
	ex = cmd.Run() == nil //it exists if we have no error
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
