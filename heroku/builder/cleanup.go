package main

type cleaner struct {
	in chan func()
	fs chan []func()
}

var cleanup = cleaner{
	in: make(chan func()),
	fs: make(chan []func()),
}

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

func (c cleaner) attach(f func()) {
	c.in <- f
}

func (c cleaner) funcs() []func() {
	return <-c.fs
}

func (c cleaner) cleanup() {
	for _, f := range c.funcs() {
		f()
	}
}
