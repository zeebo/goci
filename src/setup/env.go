package setup

import fp "path/filepath"

var (
	appdir  = env("APPROOT", ".")
	DISTDIR = fp.Join(appdir, "dist")
	VENVDIR = fp.Join(appdir, "venv")
	GOROOT  = fp.Join(appdir, "go")
)
