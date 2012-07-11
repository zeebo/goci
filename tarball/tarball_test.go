// +build !goci

package tarball

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestRealChain(t *testing.T) {
	tarball := "../foo.tar.gz"
	f, err := world.Create(tarball)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tarball)

	if err := Compress(".", f); err != nil {
		t.Fatal(err)
	}
	f.Close()

	dir, err := ioutil.TempDir("", "extract")
	if err != nil {
		t.Fatal(err)
	}
	// defer os.RemoveAll(dir)
	t.Log(dir)

	f2, err := world.Open(tarball)
	if err != nil {
		t.Fatal(err)
	}
	if err := Extract(f2, dir); err != nil {
		t.Fatal(err)
	}
}
