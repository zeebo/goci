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
	proc = p.UPID
	return
}

func cull_runner(id, proc string) func() {
	return func() {
		active_tests_lock.RLock()
		defer active_tests_lock.RUnlock()

		if _, ex := active_tests[id]; !ex {
			return
		}

		error_test(id, "timeout")

		//search for the process with the given UPID
		procs, err := hclient.List()
		if err != nil {
			log.Printf("%s error culling (list) %s: %s", id, proc, err)
			return
		}

		var pid string
		for _, p := range procs {
			if p.UPID == proc {
				pid = p.Process
				break
			}
		}

		if pid == "" {
			log.Printf("%s couldn't cull %s: process not found", id, proc)
			return
		}

		err = hclient.Kill(pid)
		if err != nil {
			log.Printf("%s error culling (kill) %s: %s", id, proc, err)
			return
		}
	}
}

func error_test(id, msg string) {
	active_tests_lock.RLock()
	defer active_tests_lock.RUnlock()

	log.Println(id, msg)

	test := active_tests[id]
	test.Error = msg
	test.Finish()

	test_complete <- id
}

func run_run_scheduler() {
	host := env("HOST", "localhost:"+env("PORT", "9080"))
	for id := range schedule_run {
		req := build_runner_url(host, id)
		cmd := fmt.Sprintf("bin/runner %s", req)
		proc, err := spawn_runner(cmd)
		if err != nil {
			error_test(id, fmt.Sprintf("error spawning %s", err))
			continue
		}
		log.Printf("%s spawned %s", id, proc)
		time.AfterFunc(90*time.Second, cull_runner(id, proc))
	}
}
