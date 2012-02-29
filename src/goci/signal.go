package main

import (
	"encoding/json"
	"time"
)

type Signal map[string]interface{}

//return a string represntation as a json encoded map
func (s Signal) String() string {
	val, _ := json.Marshal(s)
	return string(val)
}

type SignalPipe chan Signal

var (
	signalRegister   = make(chan SignalPipe) //buffered for init
	signalUnregister = make(chan SignalPipe) //buffered for init
	notify           = make(SignalPipe)
)

func init() {
	//broadcast!
	go signalBroadcast()
}

func signalBroadcast() {
	//a set of SignalPipes
	pipes := map[SignalPipe]struct{}{}
	for {
		select {
		case pipe := <-signalRegister:
			pipes[pipe] = struct{}{}
		case pipe := <-signalUnregister:
			delete(pipes, pipe)
		case signal := <-notify:
			for pipe := range pipes {
				go timeoutSend(pipe, signal, 1*time.Second)
			}
		}
	}
}

func timeoutSend(pipe SignalPipe, signal Signal, timeout time.Duration) {
	select {
	case pipe <- signal:
	case <-time.After(timeout):
	}
}
