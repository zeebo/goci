package main

import (
	"builder"
	"log"
	"setup"
	"sync"
)

//run the setup to run the work
func setup() {
	//ensure we have the go tool and vcs in parallel
	var group sync.WaitGroup
	group.Add(2)

	//check the go tool
	go func() {
		if err := setup.EnsureTool(); err != nil {
			log.Fatal(err)
		}
		group.Done()
	}()

	//check for hg + bzr + git
	go func() {
		if err := setup.EnsureVCS(); err != nil {
			log.Fatal(err)
		}
		group.Done()
	}()

	group.Wait()
	go runQueue()
}
