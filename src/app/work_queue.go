package main

import (
	"builder"
	"log"
)

const queue_size = 100

var work_queue = make(chan builder.Work, queue_size)

func run_work_queue() {
	for work := range work_queue {
		log.Println("got work item:", work.RepoPath())
		w := new_work(work)

		//create the builds for the work
		builds, err := builder.CreateBuilds(work)
		if err != nil {
			w.Error = err.Error()
			save_item <- w
			continue
		}
		save_item <- w

		//build the work struct out to include all the tests
		for _, build := range builds {
			b := new_build(build, w)
			if err := build.Error(); err != nil {
				b.Error = err.Error()
				b.cleanup(0)
				save_item <- b
				continue
			}
			save_item <- b

			paths := build.Paths()
			for _, path := range paths {
				t := new_test(path, b, w)
				schedule_test <- t
			}

			//launch a goroutine to handle cleanup after x tests have run
			go b.cleanup(len(paths))
		}
	}
}
