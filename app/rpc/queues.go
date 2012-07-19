package rpc

import "net/http"

//queue is a general purpose queue of interface items
type queue struct {
	in  chan interface{}
	out chan interface{}
}

//run handles the input and output of the queue
func (q queue) run() {
	items := make([]interface{}, 0)
	var out chan<- interface{}
	var item interface{}

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
		}
	}
}

//push puts an item in to the queue.
func (q queue) push(w interface{}) {
	q.in <- w
}

//pop pulls an item from the queue
func (q queue) pop() (w interface{}) {
	w = <-q.out
	return
}

func newQueue() (q queue) {
	q = queue{
		in:  make(chan interface{}),
		out: make(chan interface{}),
	}
	go q.run()
	return
}

//BuilderQueue is a queue with unlimited buffer for BuilderTask items.
type BuilderQueue struct {
	queue
}

//Push is an RPC method for pushing things onto the queue.
func (q BuilderQueue) Push(req *http.Request, work *BuilderTask, void *None) (err error) {
	q.push(*work)
	return
}

//Pop grabs an item from the queue.
func (q BuilderQueue) Pop() (w BuilderTask) {
	w = q.pop().(BuilderTask)
	return
}

//NewBuilderQueue creates a new queue.
func NewBuilderQueue() (q BuilderQueue) {
	q.queue = newQueue()
	return
}

//RunnerQueue is a queue with an unlimited buffer of RunnerTask items.
type RunnerQueue struct {
	queue
}

//Push is an rpc method for putting a task into the Runner queue.
func (q RunnerQueue) Push(req *http.Request, work *RunnerTask, void *None) (err error) {
	q.push(*work)
	return
}

//Pop grabs an item from the queue.
func (q RunnerQueue) Pop() (w RunnerTask) {
	w = q.pop().(RunnerTask)
	return
}

//NewRunnerQueue returns a new RunnerQueue ready to use
func NewRunnerQueue() (q RunnerQueue) {
	q.queue = newQueue()
	return
}
