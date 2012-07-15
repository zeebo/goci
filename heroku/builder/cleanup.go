package main

import "sync"

type cleaner struct {
	in chan func()
	fs chan []func()
	o  *sync.Once
}

var cleanup = cleaner{
	in: make(chan func()),
	fs: make(chan []func()),
	o:  new(sync.Once),
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
	c.o.Do(func() {
		for _, f := range c.funcs() {
			f()
		}
	})
}
