package main

import (
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/push", handlePush)
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		errLogger.Fatal(err)
	}
}
