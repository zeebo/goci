package main

import (
	"builder"
	"time"
)

type Work struct {
	Work builder.Work
	When time.Time
}

const queueSize = 100

var queue = make(chan *Work, queueSize)

func enqueue(w *Work) (ok bool) {
	select {
	case queue <- w:
		ok = true
	default:
	}
	return
}

func runQueue() {
	for {
		item := <-queue
		_ = item
	}
}

func init() {
	go run_setup()
}
