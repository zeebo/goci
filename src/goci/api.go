package main

import (
	"net/http"
	"os"
	"strings"
)

func init() {
	go ircbotLogger()
}

func postEvent(path string, sig Signal) {
	_, err := http.Post(path, "application/json", strings.NewReader(sig.String()))
	if err != nil {
		errLogger.Println(err)
	}
}

func ircbotLogger() {
	path := os.Getenv("IRC_HOOK")
	pipe := make(SignalPipe)
	signalRegister <- pipe
	for {
		sig := <-pipe

		if _, ex := sig["event"]; !ex {
			continue
		}

		switch sig["event"] {
		case "pass", "fail":
			postEvent(path, sig)
		}
	}
}
