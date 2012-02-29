package main

import "time"

type Signal map[string]interface{}

type SignalPipe chan Signal

var (
	signalRegister   = make(chan SignalPipe)
	signalUnregister = make(chan SignalPipe)
	notify           = make(SignalPipe)
)

func init() {
	//wait for env to finish init
	_ = envInit.Value()

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
				go timeoutSend(pipe, signal, 5*time.Second)
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
