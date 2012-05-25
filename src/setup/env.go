package setup

import fp "path/filepath"

var (
	appdir  = env("APPDIR", ".")
	distdir = fp.Join(appdir, "dist")
	venvdir = fp.Join(appdir, "venv")
	goroot  = fp.Join(appdir, "go")
)
