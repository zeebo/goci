package main

import (
	"fmt"
	"log"
	"time"
)

var (
	schedule_run = make(chan string)
)

func spawn_runner(cmd string) (proc string, err error) {
	p, err := hclient.Run(cmd)
	if err != nil {
		return
	}
	log.Printf("running %s on %s", cmd, p)
	proc = p.Process
	return
}

func cull_runner(id, proc string) func() {
	return func() {
		active_tests_lock.RLock()
		defer active_tests_lock.RUnlock()

		test, ex := active_tests[id]
		if !ex {
			return
		}

		log.Println(id, "timeout")

		test.Error = "timeout"
		test.Finish()

		err := hclient.Kill(proc)
		if err != nil {
			log.Printf("error culling %s[%s]: %s", proc, id, err)
		}

		test_complete <- id
	}
}

func run_run_scheduler() {
	host := env("HOST", "localhost:"+env("PORT", "9080"))
	for id := range schedule_run {
		req := build_runner_url(host, id)
		cmd := fmt.Sprintf("bin/runner %s", req)
		proc, err := spawn_runner(cmd)
		if err != nil {
			msg := fmt.Sprintf("error spawning %s: %s", id, err)
			log.Printf(msg)
			active_tests_lock.RLock()
			defer active_tests_lock.RUnlock()

			test := active_tests[id]
			test.Error = msg
			test.Finish()

			test_complete <- id
			return
		}
		log.Printf("spawned %s for %s", proc, id)
		time.AfterFunc(90*time.Second, cull_runner(id, proc))
	}
}
