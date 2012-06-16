package worker

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

var (
	schedule_test     = make(chan *Test)
	buffer_test       = make(chan string)
	num_tests         = make(chan bool, 1)   //simultaneous running tests
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

var testNotFound = errors.New("test not found")

func Serve(w io.Writer, id string) (err error) {
	active_tests_lock.RLock()
	defer active_tests_lock.RUnlock()

	test, ex := active_tests[id]
	if !ex {
		err = testNotFound
		return
	}
	f, err := os.Open(test.Path)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	if err != nil {
		return
	}
	test.Start()
	return
}

func Response(r io.Reader, id string) (err error) {
	active_tests_lock.RLock()
	defer active_tests_lock.RUnlock()

	test, ex := active_tests[id]
	if !ex {
		err = testNotFound
		return
	}
	by, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}
	s := string(by)
	test.Output = s
	test.Passed = strings.HasSuffix(s, "\nPASS\n")
	test.Finish()
	test_complete <- id
	return
}

func Error(r io.Reader, id string) (err error) {
	active_tests_lock.RLock()
	defer active_tests_lock.RUnlock()

	test, ex := active_tests[id]
	if !ex {
		err = testNotFound
		return
	}
	by, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}
	test.Error = string(by)
	test.Finish()
	test_complete <- id
	return
}
