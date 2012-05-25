package setup

import (
	"log"
	fp "path/filepath"
)

var (
	appdir  = env("APPROOT", ".")
	DISTDIR = fp.Join(appdir, "dist")
	VENVDIR = fp.Join(appdir, "venv")
	GOROOT  = fp.Join(appdir, "go")
)

func init() {
	log.Println("Setup initialized with:")
	log.Println("\tDISTDIR: ", DISTDIR)
	log.Println("\tVENVDIR: ", VENVDIR)
	log.Println("\tGOROOT:  ", GOROOT)
}
