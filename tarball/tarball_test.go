// +build !goci

package tarball

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestRealChain(t *testing.T) {
	tarball := "../foo.tar.gz"
	defer os.Remove(tarball)
	if err := CompressFile(".", tarball); err != nil {
		t.Fatal(err)
	}

	dir, err := ioutil.TempDir("", "extract")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := ExtractFile(tarball, dir); err != nil {
		t.Fatal(err)
	}
}
