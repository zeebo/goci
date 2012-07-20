// +build !goci

package builder

import (
	"github.com/zeebo/goci/app/rpc"
	"testing"
)

func TestSingleImport(t *testing.T) {
	w := rpc.Work{
		Revision:   "d80b78af7cba69b8a152a46b8a1f7b9f72d954a0",
		ImportPath: "github.com/zeebo/irc",
	}
	b := New("", "", "")
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
	w := rpc.Work{
		Revision:    "cc5e03949586c5b697c9a3080cc3bf7501a14d96",
		ImportPath:  "github.com/dustin/githubhooks",
		Subpackages: true,
	}
	b := New("", "", "")
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

func TestBazaarImport(t *testing.T) {
	w := rpc.Work{
		Revision:   "140",
		ImportPath: "labix.org/v2/mgo",
	}
	b := New("", "", "")
	bs, _, err := b.Build(&w)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", bs)
	for _, bu := range bs {
		bu.Clean()
	}
}
