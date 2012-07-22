package main

import (
	"fmt"
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/heroku"
	"log"
	"time"
)

//defaultClient is a heroku client for managing spawning requests
var defaultClient = heroku.NewManaged(
	env("APP_NAME", "goci"), //name of the app (goci)
	env("API_KEY", "fail"),  //api key (secret!)
	2,                       // maximum of 2 dynos running tests
	1*time.Minute,           //can only live for 1 minute
)

func process_run(task rpc.RunnerTask) {
	log.Printf("Running task: %+v", task)

	//create a runner task for the incoming task
	r := &runnerTask{
		task:  task,
		resps: make(chan rpc.Output, len(task.Tests)),
		ids:   make(map[string]chan string),
	}
	go r.run()

	//register our task with our task map
	defaultTaskMap.Register(r)

	//create the rpc url
	rpcUrl := urlWithPath(reverse("rpc"))
	for i, rt := range task.Tests {

		//create an action for our managed heroku client
		action := heroku.Action{
			Command: fmt.Sprintf("bin/runner %s %s %d", rpcUrl, task.ID, i),
			Error: func(err string) {
				r.resps <- rpc.Output{
					ImportPath: rt.ImportPath,
					Config:     rt.Config,
					Error:      err,
				}
			},
		}

		//create the channel for our import path id and add it to the map
		ch := make(chan string, 1)
		r.ids[rt.ImportPath] = ch

		//run the action
		id, err := defaultClient.Run(action)
		if err != nil {
			log.Printf("Error spawning run task: %v", err)
			return
		}

		//send the id down the channel for the runner
		ch <- id
	}
}
