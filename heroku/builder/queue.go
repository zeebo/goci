package main

import (
	"github.com/zeebo/goci/app/rpc"
	"net/http"
)

//BuilderQueue is a queue with unlimited buffer for work items.
type BuilderQueue struct {
	in  chan rpc.BuilderTask
	out chan rpc.BuilderTask
	ich chan []rpc.BuilderTask
}

//run performs the logic of the queue through the channels.
func (q BuilderQueue) run() {
	items := make([]rpc.BuilderTask, 0)
	var out chan<- rpc.BuilderTask
	var item rpc.BuilderTask

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
func (q BuilderQueue) Push(req *http.Request, work *rpc.BuilderTask, void *rpc.None) (err error) {
	q.push(*work)
	return
}

//Items is an RPC method for getting the current items in the queue.
func (q BuilderQueue) Items(req *http.Request, void *rpc.None, resp *[]rpc.BuilderTask) (err error) {
	*resp = q.items()
	return
}

//push puts an item in to the queue.
func (q BuilderQueue) push(w rpc.BuilderTask) {
	q.in <- w
}

//pop grabs an item from the queue.
func (q BuilderQueue) pop() (w rpc.BuilderTask) {
	w = <-q.out
	return
}

//items returns the current set of queued items.
func (q BuilderQueue) items() []rpc.BuilderTask {
	return <-q.ich
}

//create our local queue.
var builderQueue = newBuilderQueue()

//newBuilderQueue creates a new queue.
func newBuilderQueue() (q BuilderQueue) {
	q = BuilderQueue{
		in:  make(chan rpc.BuilderTask),
		out: make(chan rpc.BuilderTask),
		ich: make(chan []rpc.BuilderTask),
	}
	go q.run()
	return
}

//register our builderQueue service.
func init() {
	if err := rpcServer.RegisterService(builderQueue, ""); err != nil {
		bail(err)
	}
}
