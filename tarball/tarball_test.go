package tarball

import (
	"os"
	"testing"
)

func TestRealCompress(t *testing.T) {
	defer os.Remove("foo.tar.gz")
	if err := Compress(".", "foo.tar.gz"); err != nil {
		t.Fatal(err)
	}
}
