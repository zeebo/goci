package main

import "os"

//setup so the tests run locally
func init() {
	envInit.Wait()
	defer logger.Println("Darwin environment setup finished.")

	goHost = `darwin-amd64`

	//set the path for local testing
	os.Setenv("PATH", "/usr/bin:/bin")
}
