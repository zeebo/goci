package main

import (
	"builder"
	"log"
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
		res, err := builder.Run(item.Work)

		//do something with the result and error
		log.Println(err, res)
	}
}

func init() {
	go run_setup()
}
