package web

import (
	"fmt"
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/heroku"
)

func (r *Runner) process(task rpc.RunnerTask) {
	//create a runner task for the incoming task
	rtask := &runnerTask{
		mc:    r.mc,
		tm:    r.tm,
		task:  task,
		resps: make(chan rpc.Output, len(task.Tests)),
		ids:   make(map[string]chan string),
	}
	go rtask.run()

	//register our task with our task map
	r.tm.Register(rtask)

	//create the rpc url
	for i, rt := range task.Tests {

		//create an action for our managed heroku client
		action := heroku.Action{
			Command: fmt.Sprintf("bin/runner %s %s %d", r.base, task.ID, i),
			Error: func(err string) {
				rtask.resps <- rpc.Output{
					ImportPath: rt.ImportPath,
					Config:     rt.Config,
					Error:      err,
				}
			},
		}

		//create the channel for our import path id and add it to the map
		ch := make(chan string, 1)
		rtask.ids[rt.ImportPath] = ch

		//run the action
		id, err := r.mc.Run(action)
		if err != nil {
			return
		}

		//send the id down the channel for the runner
		ch <- id
	}
}
