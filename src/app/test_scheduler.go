package main

import "sync"

var (
	schedule_test     = make(chan *Test)
	buffer_test       = make(chan string)
	num_tests         = make(chan bool, 1)
	test_complete     = make(chan string, 1) //needs to buffer to avoid deadlock on the active test mutex
	active_tests      = make(map[string]*Test)
	active_tests_lock sync.RWMutex
)

func run_test_scheduler() {
	go run_test_buffer()
	for {
		select {
		case t := <-schedule_test:
			schedule(t)
		case id := <-test_complete:
			unschedule(id)
		}
	}
}

func schedule(t *Test) {
	active_tests_lock.Lock()
	defer active_tests_lock.Unlock()

	id := t.WholeID()
	active_tests[id] = t

	buffer_test <- id
}

func unschedule(id string) {
	active_tests_lock.Lock()
	defer active_tests_lock.Unlock()

	<-num_tests
	t, ok := active_tests[id]
	if ok {
		delete(active_tests, id)
		save_item <- t
	}
}

func run_test_buffer() {
	var buffer []string
	for {
		if len(buffer) > 0 {
			select {
			case id := <-buffer_test:
				buffer = append(buffer, id)
			case num_tests <- true:
				schedule_run <- buffer[0]
				buffer = buffer[1:]
			}
		} else {
			buffer = append(buffer, <-buffer_test)
		}
	}
}
