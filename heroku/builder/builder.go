package main

import (
	"log"
	"net/http"
	"os"
)

func env(key, def string) (r string) {
	if r = os.Getenv(key); r == "" {
		r = def
	}
	return
}

func main() {
	go func() {
		if err := setup(); err != nil {
			log.Panic(err)
		}
		announce()
	}()
	defer cleanup.cleanup()

	panic(http.ListenAndServe(":9080", nil))
}
