// +build !goci

package setup

import "testing"

func TestInstall(t *testing.T) {
	_, und := testMode(t)
	defer und()

	if err := InstallGo("/tmp/goroot/go"); err != nil {
		t.Fatal(err)
	}
}
