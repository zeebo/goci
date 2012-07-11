package tarball

import (
	"os"
	"testing"
)

func TestCompress(t *testing.T) {
	defer func(e localWorld) { world = e }(world)
	world = newTestWorld(t, 3)

	f, err := world.Create("foo.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	if err := Compress("0tarball", f); err != nil {
		t.Fatal(err)
	}
}

func TestCompressNotDirectory(t *testing.T) {
	defer func(e localWorld) { world = e }(world)
	world = newTestWorld(t, 3)

	f, err := world.Create("foo.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	if err := Compress("tarball", f); err != nil {
		t.Fatal(err)
	}
}

func TestExtract(t *testing.T) {
	defer func(e localWorld) { world = e }(world)
	tw := newTestWorld(t, 3)
	r, err := os.Open("tarb.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	tw.files["foo"] = r
	world = tw

	f, err := world.Open("foo")
	if err != nil {
		t.Fatal(err)
	}

	if err := Extract(f, "bar"); err != nil {
		t.Fatal(err)
	}
}
