package setup

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"
)

func TestDownload(t *testing.T) {
	//only works on linux
	if runtime.GOOS != "linux" {
		t.Log("test skipped: only runs on linux")
		return
	}

	dir, err := ioutil.TempDir("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := download(dir); err != nil {
		t.Fatal(err)
	}

	if !toolExists(dir) {
		t.Fatal("No tool found")
	}
}
