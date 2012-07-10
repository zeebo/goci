package tarball

import "testing"

func TestCompress(t *testing.T) {
	defer func(e localWorld) { world = e }(world)
	world = newTestWorld(t, 3)

	if err := Compress("0tarball", "foo.tar.gz"); err != nil {
		t.Fatal(err)
	}
}

func TestCompressNotDirectory(t *testing.T) {
	defer func(e localWorld) { world = e }(world)
	world = newTestWorld(t, 3)

	if err := Compress("tarball", "foo.tar.gz"); err != nil {
		t.Fatal(err)
	}
}
