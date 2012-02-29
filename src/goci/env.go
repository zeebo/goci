package main

import (
	"fmt"
	"os"
	"runtime"
)

var (
	cacheDir  = os.TempDir()
	goVersion = `weekly.2012-02-22`
	goHost    = fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
)
