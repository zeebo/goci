package tarball

import (
	// "io/ioutil"
	"os"
	"testing"
)

func TestRealCompress(t *testing.T) {
	defer os.Remove("foo.tar.gz")
	if err := Compress(".", "foo.tar.gz"); err != nil {
		t.Fatal(err)
	}
}

// func TestRealExtract(t *testing.T) {
// 	dir, err := ioutil.TempDir("", "extract")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(dir)

// 	if err := Extract("../foo.tar.gz", dir); err != nil {
// 		t.Fatal(err)
// 	}
// }
