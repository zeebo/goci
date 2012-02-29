package main

func init() {
	go notifyLogger()
}

func notifyLogger() {
	pipe := make(SignalPipe)
	signalRegister <- pipe
	for {
		logger.Println("signal:", <-pipe)
	}
}
