package main

import (
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/builder"
	"net/http"
)

type Task struct {
	Work   builder.Work
	Runner string
}

//TaskQueue is a queue with unlimited buffer for work items
type TaskQueue struct {
	in  chan Task
	out chan Task
	ich chan []Task
}

//run performs the logic of the queue through the channels
func (q TaskQueue) run() {
	items := make([]Task, 0)
	var out chan<- Task
	var item Task

	for {
		select {
		case w := <-q.in:
			items = append(items, w)
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
		case q.ich <- items:
		}
	}
}

//Queue is an RPC method for pushing things onto the queue.
func (q TaskQueue) Push(req *http.Request, work *Task, void *rpc.None) (err error) {
	log.Println("Pushing", *work)
	q.push(*work)
	return
}

//Items is an RPC method for getting the current items in the queue.
func (q TaskQueue) Items(req *http.Request, void *None, resp *[]Task) (err error) {
	log.Println("Getting Items")
	*resp = q.items()
	return
}

//push puts an item in to the queue.
func (q TaskQueue) push(w Task) {
	q.in <- w
}

//pop grabs an item from the queue.
func (q TaskQueue) pop() (w Task) {
	w = <-q.out
	return
}

//items returns the current set of queued items.
func (q TaskQueue) items() []Task {
	return <-q.ich
}

//create our local queue
var queue = TaskQueue{
	in:  make(chan Task),
	out: make(chan Task),
	ich: make(chan []Task),
}

func init() {
	go queue.run()
	if err := rpcServer.RegisterService(queue, "Queue"); err != nil {
		bail(err)
	}
}
