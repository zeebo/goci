package main

import (
	"github.com/zeebo/goci/app/rpc"
	"net/http"
)

//RunnerQueue is a queue with an unlimited buffer of runner items
type RunnerQueue struct {
	in  chan rpc.RunnerTask
	out chan rpc.RunnerTask
}

//run performs the logic of the queue through the channels
func (q RunnerQueue) run() {
	items := make([]rpc.RunnerTask, 0)
	var out chan<- rpc.RunnerTask
	var item rpc.RunnerTask

	for {
		select {
		case t := <-q.in:
			items = append(items, t)
			if out == nil {
				out = q.out
				item = items[0]
			}
		case out <- item:
			items = items[1:]
			if len(items) == 0 {
				out = nil
			} else {
				item = items[0]
			}
		}
	}
}

//Push is an rpc method for putting a task into the Runner queue.
func (q RunnerQueue) Push(req *http.Request, work *rpc.RunnerTask, void *rpc.None) (err error) {
	q.push(*work)
	return
}

//push is an internal method for pushing tasks.
func (q RunnerQueue) push(w rpc.RunnerTask) {
	q.in <- w
}

//pop is an internal method for getting a task from the queue.
func (q RunnerQueue) pop() (w rpc.RunnerTask) {
	w = <-q.out
	return
}

//runnerQueue is the queue we use in our program.
var runnerQueue = newRunnerQueue()

//newRunnerQueue returns a new RunnerQueue ready to use
func newRunnerQueue() (q RunnerQueue) {
	q = RunnerQueue{
		in:  make(chan rpc.RunnerTask),
		out: make(chan rpc.RunnerTask),
	}
	go q.run()
	return
}

//register our runnerQueue service
func init() {
	if err := rpcServer.RegisterService(runnerQueue, ""); err != nil {
		bail(err)
	}
}
