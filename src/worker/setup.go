package worker

import (
	"builder"
	"heroku"
	"log"
	"setup"
	"sync"
)

//all the config we need
var (
	debug   = false
	hclient *heroku.Client
	host    string
	goroot  string
)

//run the setup to run the work
func run_setup() {
	//set the goroot env var to the default in the env
	setup.GOROOT, builder.GOROOT = goroot, goroot

	log.Println("running setup...")
	setup.PrintVars()

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
	change_state <- StateIdle
	go run_work_queue()
}
