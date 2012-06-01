package setup

import (
	"log"
	"os"
	fp "path/filepath"
)

var (
	appdir  = env("APPROOT", ".")
	DISTDIR = fp.Join(appdir, "dist")
	VENVDIR = fp.Join(appdir, "venv")
	GOROOT  = fp.Join(appdir, "go")
)

func PrintVars() {
	log.Println("\tDISTDIR: ", DISTDIR)
	log.Println("\tVENVDIR: ", VENVDIR)
	log.Println("\tGOROOT:  ", GOROOT)
	wd, _ := os.Getwd()
	log.Println("\tWD:      ", wd)
}
