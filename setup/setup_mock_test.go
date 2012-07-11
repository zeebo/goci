package setup

import (
	"github.com/zeebo/goci/environ"
	"github.com/zeebo/goci/tarball"
	"reflect"
	"testing"
)

func testMode(t *testing.T) (environ.TestEnv, func()) {
	prevWorld, prevTbWorld := World, tarball.World
	tw := environ.NewTest(t, 3)
	World, tarball.World = tw, tw
	return tw, func() {
		World, tarball.World = prevWorld, prevTbWorld
	}
}

func compare(t *testing.T, expect, got []string) {
	if reflect.DeepEqual(expect, got) {
		return
	}

	t.Fail()
	t.Log("Expected")
	for _, ev := range expect {
		t.Logf("\t%q", ev)
	}
	t.Log("Got")
	for _, ev := range got {
		t.Logf("\t%q", ev)
	}
}

func TestEndsInGo(t *testing.T) {
	tw, und := testMode(t)
	defer und()

	if err := InstallGo("/tmp"); err != ErrInvalidGOROOT {
		t.Fatal("Expected an error. Got %v", err)
	}

	//nothing should happen because directory is bad
	expect := []string{}

	compare(t, expect, tw.Events())
}
