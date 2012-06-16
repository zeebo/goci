package worker

import "sync"

var (
	schedule_test = make(chan *Test)
	buffer_test   = make(chan string)
	num_tests     = make(chan bool, 1)
	test_complete = make(chan string, 1) //needs to buffer to avoid deadlock on the active test mutex
	active_tests  = make(map[string]*Test)

	TestLock sync.RWMutex
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
	TestLock.Lock()
	defer TestLock.Unlock()

	id := t.WholeID()
	active_tests[id] = t

	buffer_test <- id
}

func unschedule(id string) {
	TestLock.Lock()
	defer TestLock.Unlock()

	<-num_tests
	delete(active_tests, id)
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

func GetTest(id string) (test *Test, ex bool) {
	test, ex = active_tests[id]
	return
}

func Complete(test *Test) {
	test_complete <- test.ID
}
