package main

import (
	"builder"
	"log"
)

const queue_size = 100

var work_queue = make(chan builder.Work, queue_size)

func run_work_queue() {
	done := make(chan bool)
	for work := range work_queue {
		change_state <- StateRunning

		log.Println("got work item:", work.RepoPath())
		w := new_work(work)

		//create the builds for the work
		builds, err := builder.CreateBuilds(work)
		if err != nil {
			w.Error = err.Error()
			w.cleanup(0, nil)
			continue
		}

		go w.cleanup(len(builds), done)

		//build the work struct out to include all the tests
		for _, build := range builds {
			b := new_build(build, w)
			if err := build.Error(); err != nil {
				b.Error = err.Error()
				b.Passed = false
				b.cleanup(0)
				continue
			}

			paths := build.Paths()
			go b.cleanup(len(paths))
			for _, path := range paths {
				t := new_test(path, b, w)
				schedule_test <- t
			}
		}

		//wait for the work to finish
		<-done
		change_state <- StateIdle
	}
}
