package tarball

import (
	"compress/gzip"
	"os"
	"reflect"
	"testing"
)

func compare(t *testing.T, expect, got []string) {
	if reflect.DeepEqual(expect, got) {
		return
	}

	t.Fail()
	t.Log("Expected")
	for _, ev := range expect {
		t.Logf("\t%q", ev)
	}
}

func TestCompress(t *testing.T) {
	defer func(e localWorld) { world = e }(world)
	tw := newTestWorld(t, 3)
	world = tw

	//disable compression for deterministic output
	defer func(c int) { compression = c }(compression)
	compression = gzip.NoCompression

	if err := CompressFile("0tarball", "foo.tar.gz"); err != nil {
		t.Fatal(err)
	}

	expect := []string{
		"world: Create(foo.tar.gz, 0666)",
		"world: Stat(0tarball): 0tarball:0",
		"0tarball: Mode(): dir:true",
		"0tarball: Mode(): dir:true",
		"foo.tar.gz: Write(10)",
		"world: Readdir(0tarball): [15003edb:0 75f91b0b:983]",
		"15003edb: Mode(): dir:true",
		"15003edb: Mode(): dir:true",
		"world: Readdir(0tarball/15003edb): [7fabe9ae:261]",
		"7fabe9ae: Mode(): dir:false",
		"7fabe9ae: Mode(): dir:false",
		"7fabe9ae: Size(): 261",
		"world: Open(0tarball/15003edb/7fabe9ae)",
		"0tarball/15003edb/7fabe9ae: Read(261)",
		"0tarball/15003edb/7fabe9ae: Close()",
		"75f91b0b: Mode(): dir:false",
		"75f91b0b: Mode(): dir:false",
		"75f91b0b: Size(): 983",
		"world: Open(0tarball/75f91b0b)",
		"0tarball/75f91b0b: Read(983)",
		"0tarball/75f91b0b: Close()",
		"foo.tar.gz: Write(1)",
		"foo.tar.gz: Write(4)",
		"foo.tar.gz: Write(4608)",
		"foo.tar.gz: Write(1)",
		"foo.tar.gz: Write(4)",
		"foo.tar.gz: Write(8)",
		"foo.tar.gz: Close()",
	}

	compare(t, expect, tw.events())
}

func TestCompressNotDirectory(t *testing.T) {
	defer func(e localWorld) { world = e }(world)
	tw := newTestWorld(t, 3)
	world = tw

	//disable compression for deterministic output
	defer func(c int) { compression = c }(compression)
	compression = gzip.NoCompression

	if err := CompressFile("tarball", "foo.tar.gz"); err != nil {
		t.Fatal(err)
	}

	expect := []string{
		"world: Create(foo.tar.gz, 0666)",
		"world: Stat(tarball): tarball:861",
		"tarball: Mode(): dir:false",
		"tarball: Mode(): dir:false",
		"tarball: Size(): 861",
		"foo.tar.gz: Write(10)",
		"world: Open(tarball)",
		"tarball: Read(861)",
		"tarball: Close()",
		"foo.tar.gz: Write(1)",
		"foo.tar.gz: Write(4)",
		"foo.tar.gz: Write(2560)",
		"foo.tar.gz: Write(1)",
		"foo.tar.gz: Write(4)",
		"foo.tar.gz: Write(8)",
		"foo.tar.gz: Close()",
	}

	compare(t, expect, tw.events())
}

func TestExtract(t *testing.T) {
	defer func(e localWorld) { world = e }(world)
	tw := newTestWorld(t, 3)

	//do a for realsies open and throw it into the world
	r, err := os.Open("tarb.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	tw.files["foo"] = r
	world = tw

	if err := ExtractFile("foo", "bar"); err != nil {
		t.Fatal(err)
	}

	expect := []string{
		"world: Open(foo)",
		"world: returned set file",
		"world: MkdirAll(bar, 0755)",
		"world: Create(bar/mock_test.go, 0644)",
		"bar/mock_test.go: Write(2618)",
		"bar/mock_test.go: Close()",
		"world: Create(bar/tarball.go, 0644)",
		"bar/tarball.go: Write(3870)",
		"bar/tarball.go: Close()",
		"world: Create(bar/tarball_mock_test.go, 0644)",
		"bar/tarball_mock_test.go: Write(784)",
		"bar/tarball_mock_test.go: Close()",
		"world: Create(bar/tarball_test.go, 0644)",
		"bar/tarball_test.go: Write(490)",
		"bar/tarball_test.go: Close()",
		"world: Create(bar/walk.go, 0644)",
		"bar/walk.go: Write(1487)",
		"bar/walk.go: Close()",
	}

	compare(t, expect, tw.events())
}
