package main

var notifyLoggerPipe = make(SignalPipe)

func init() {
	//async send to register the pipe
	go func() {
		signalRegister <- notifyLoggerPipe
	}()
	go notifyLogger()
}

func notifyLogger() {
	for {
		sig := <-notifyLoggerPipe
		logger.Println("signal:", sig)
	}
}
