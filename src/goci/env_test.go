package main

import "os"

//setup so the tests run locally
func init() {
	envInit.Wait()

	//change the cacheDir to something we can actually use
	cacheDir = os.TempDir()
	goHost = `darwin-amd64`

	//set the path for local testing
	os.Setenv("PATH", "/usr/bin:/bin")
}
