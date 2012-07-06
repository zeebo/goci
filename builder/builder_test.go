// +build !goci

package builder

import "testing"

func TestSingleImport(t *testing.T) {
	w := Work{
		Revision:   "e9dd26552f10d390b5f9f59c6a9cfdc30ed1431c",
		ImportPath: "github.com/zeebo/irc",
	}
	b := New("", "", "", "")
	bs, _, err := b.Build(&w)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", bs)
	for _, bu := range bs {
		bu.Clean()
	}
}

func TestSubpackageImport(t *testing.T) {
	w := Work{
		Revision:    "cc5e03949586c5b697c9a3080cc3bf7501a14d96",
		ImportPath:  "github.com/dustin/githubhooks",
		Subpackages: true,
	}
	b := New("", "", "", "")
	bs, _, err := b.Build(&w)
	if err != nil {
		t.Fatal(err)
	}
	if len(bs) != 2 {
		t.Errorf("Got %d builds. Expected 2", len(bs))
	}
	t.Logf("%+v", bs)
	for _, bu := range bs {
		bu.Clean()
	}
}
