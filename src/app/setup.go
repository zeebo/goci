package main

import (
	"log"
	"setup"
	"sync"
)

//run the setup to run the work
func run_setup() {
	log.Println("running setup...")
	//ensure we have the go tool and vcs in parallel
	var group sync.WaitGroup
	group.Add(2)

	//check the go tool
	go func() {
		if err := setup.EnsureTool(); err != nil {
			log.Fatal(err)
		}
		log.Println("tooling complete")
		group.Done()
	}()

	//check for hg + bzr
	go func() {
		if err := setup.EnsureVCS(); err != nil {
			log.Fatal(err)
		}
		log.Println("vcs complete")
		group.Done()
	}()

	group.Wait()

	log.Println("setup complete. running queue")
	go runQueue()
}
