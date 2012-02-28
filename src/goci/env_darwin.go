package main

import "os"

//setup so the tests/app runs locally
func init() {
	envInit.Wait()
	defer logger.Println("Darwin environment setup finished.")

	goHost = `darwin-amd64`

	//set the path for local testing
	os.Setenv("PATH", "/usr/bin:/bin")
}
