package builder

import "sync"

//cleaner is a type that holds a set of functions to be run, like a defer.
type cleaner struct {
	in chan func()
	fs chan []func()
	o  *sync.Once
}

//cleanup is an instance of a cleaner.
var cleanup = newCleaner()

//newCleaner returns a new cleaner ready to use.
func newCleaner() (c cleaner) {
	c = cleaner{
		in: make(chan func()),
		fs: make(chan []func()),
		o:  new(sync.Once),
	}
	go c.run()
	return
}

//run handles the event loop for the cleaner.
func (c cleaner) run() {
	funcs := make([]func(), 0)
	for {
		select {
		case f := <-c.in:
			funcs = append(funcs, f)
		case c.fs <- funcs:
		}
	}
}

//attach takes a function and attaches it for cleanup.
func (c cleaner) attach(f func()) {
	c.in <- f
}

//funcs returns the set of functions to be run on cleanup.
func (c cleaner) funcs() []func() {
	return <-c.fs
}

//cleanup executes the cleanup of the functions once.
func (c cleaner) cleanup() {
	c.o.Do(func() {
		for _, f := range c.funcs() {
			f()
		}
	})
}
