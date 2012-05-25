package setup

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	fp "path/filepath"
	"runtime"
	"sync"
)

var toolLock sync.Mutex

func EnsureTool() (err error) {
	toolLock.Lock()
	defer toolLock.Unlock()

	//fast path: tool exists
	if toolExists() {
		return
	}

	//gotta download/unzip it
	err = toolDownload()
	return
}

func toolExists() (ex bool) {
	ex = exists("go")
	return
}

func toolDownload() (err error) {
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
	dlpath := fmt.Sprintf("https://go.googlecode.com/files/go1.0.1.%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	log.Println("downloading", dlpath, "to", tbPath)
	resp, err := http.Get(dlpath)
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

	//close the tarball to flush writes out
	if err = tarball.Close(); err != nil {
		err = fmt.Errorf("error closing tarball: %s", err)
		return
	}

	//extract the tarball
	log.Println("extracting", tbPath, "to", tmpDir)
	var tarout, tarerr bytes.Buffer
	cmd := exec.Command("tar", "zxf", tbPath)
	cmd.Dir = tmpDir
	cmd.Stdout = &tarout
	cmd.Stderr = &tarerr
	err = cmd.Run()
	log.Println("tar out:", tarout.String())
	log.Println("tar err:", tarerr.String())
	if err != nil {
		err = fmt.Errorf("error untarring file: %s", err)
		return
	}

	//make the destination directory
	if err = os.MkdirAll(GOROOT, 0777); err != nil {
		err = fmt.Errorf("error making destination directory: %s", err)
		return
	}

	//copy the files into the destination directory
	copyFiles := fp.Join(tmpDir, "go") + string(fp.Separator)
	log.Println("copying", copyFiles, "to", GOROOT)
	var cpout, cperr bytes.Buffer
	cmd = exec.Command("cp", "-a", copyFiles, GOROOT)
	cmd.Stdout = &cpout
	cmd.Stderr = &cperr
	err = cmd.Run()
	log.Println("cp out:", cpout.String())
	log.Println("cp err:", cperr.String())
	if err != nil {
		err = fmt.Errorf("error copying files to destination: %s", err)
		return
	}

	//DEBUG:run an ls of the tmpdir to see what happened and exit
	cmd = exec.Command("ls", "-al", copyFiles)
	var lsout bytes.Buffer
	cmd.Stdout = &lsout
	err = cmd.Run()
	log.Println("ls out:", lsout.String())
	log.Println("ls err:", err)

	cmd = exec.Command("ls", "-al", GOROOT)
	lsout.Reset()
	cmd.Stdout = &lsout
	err = cmd.Run()
	log.Println("ls out:", lsout.String())
	log.Println("ls err:", err)

	//check if we can run the go command.
	//if not, try adding GOROOT/bin to the path
	if !toolExists() {
		bindir := fp.Join(GOROOT, "bin")
		if _, e := os.Stat(fp.Join(bindir, "go")); e != nil {
			err = fmt.Errorf("can't find go command where it was expected: %s", e)
			return
		}
		path := fmt.Sprintf("%s:%s", os.Getenv("PATH"), bindir)
		os.Setenv("PATH", path)
	}
	//if we still don't have the tool, we have an error
	if !toolExists() {
		err = fmt.Errorf("tool downloaded but can't find go command")
	}
	return
}
