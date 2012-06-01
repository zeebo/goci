package main

import "log"

var (
	schedule_run = make(chan string)
)

func init() {
	go run_scheduler()
}

func run_scheduler() {
	host := env("HOST", "localhost:"+env("PORT", "9080"))
	for id := range schedule_run {
		req, res, err := build_url_pair(host, id)
		log.Println("bin/runner", req, res, err)
	}
}
