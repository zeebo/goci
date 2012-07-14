package main

import (
	"github.com/zeebo/goci/builder"
	"log"
	"net/http"
)

//WorkQueue is a queue with unlimited buffer for work items
type WorkQueue struct {
	in  chan builder.Work
	out chan builder.Work
	ich chan []builder.Work
}

//run performs the logic of the queue through the channels
func (q WorkQueue) run() {
	items := make([]builder.Work, 0)
	var out chan<- builder.Work
	var item builder.Work

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

//None is an empty rpc element
type None struct{}

//Queue is an RPC method for pushing things onto the queue.
func (q WorkQueue) Push(req *http.Request, work *builder.Work, void *None) (err error) {
	log.Println("Pushing", *work)
	q.push(*work)
	return
}

//Items is an RPC method for getting the current items in the queue.
func (q WorkQueue) Items(req *http.Request, void *None, resp *[]builder.Work) (err error) {
	log.Println("Getting Items")
	*resp = q.items()
	return
}

//push puts an item in to the queue.
func (q WorkQueue) push(w builder.Work) {
	q.in <- w
}

//pop grabs an item from the queue.
func (q WorkQueue) pop() (w builder.Work) {
	w = <-q.out
	return
}

//items returns the current set of queued items.
func (q WorkQueue) items() []builder.Work {
	return <-q.ich
}

//create our local queue
var queue = WorkQueue{
	in:  make(chan builder.Work),
	out: make(chan builder.Work),
	ich: make(chan []builder.Work),
}

func init() {
	go queue.run()
	if err := rpc_server.RegisterService(queue, "Queue"); err != nil {
		log.Panic(err)
	}
}
