package tarball

import (
	"os"
	"testing"
)

func TestRealCompress(t *testing.T) {
	pre := "../"
	defer os.Remove(pre + "foo.tar.gz")
	if err := Compress(pre+"tarball", pre+"foo.tar.gz"); err != nil {
		t.Fatal(err)
	}
}
