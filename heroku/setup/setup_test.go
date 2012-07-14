// +build !goci

package setup

import "testing"

func TestInstallGo(t *testing.T) {
	_, und := testMode(t)
	defer und()

	if bin, err := InstallGo("", "", "/tmp/inst"); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("bin dir: %s", bin)
	}
}

func TestInstallVCS(t *testing.T) {
	_, und := testMode(t)
	defer und()

	if bin, err := InstallVCS("/tmp/inst"); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("bin dir: %s", bin)
	}
}
