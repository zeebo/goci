package heroku

import (
	"sync"
	"time"
)

type Action struct {
	Command string
	Error   func(string)
}

type taskInfo struct {
	a    Action
	proc *Process
}

type ManagedClient struct {
	client *Client
	sem    chan bool
	ttl    time.Duration

	mu    sync.Mutex //to protect spawn
	spawn map[string]taskInfo
}

func NewManaged(app, api string, count int, ttl time.Duration) *ManagedClient {
	return &ManagedClient{
		client: New(app, api),
		sem:    make(chan bool, count),
		ttl:    ttl,
		spawn:  map[string]taskInfo{},
	}
}

func (m *ManagedClient) acquire() { m.sem <- true }
func (m *ManagedClient) release() { <-m.sem }

func (m *ManagedClient) Run(a Action) (id string, err error) {
	//acquire the semaphore
	m.acquire()

	//run the command
	p, err := m.client.Run(a.Command)
	if err != nil {
		a.Error("error spawning runner: " + err.Error())
		m.release()
		return
	}

	//add process to our spawn map
	id = p.UPID
	m.mu.Lock()
	m.spawn[id] = taskInfo{a, p}
	m.mu.Unlock()

	go m.cull(id)

	return
}

func (m *ManagedClient) Finished(id string) {
	m.mu.Lock()
	//need to check so we don't over release
	if _, ok := m.spawn[id]; ok {
		delete(m.spawn, id)
		m.release()
	}
	m.mu.Unlock()
}

func (m *ManagedClient) cull(id string) {
	//wait the ttl
	<-time.After(m.ttl)

	//grab the lock
	m.mu.Lock()
	defer m.mu.Unlock()

	//see if we have the action still
	info, ok := m.spawn[id]
	if !ok {
		return
	}
	delete(m.spawn, id)
	defer m.release() //make sure we release no matter what

	//kill the process, and panic if we don't (never leak processes!)
	if err := m.client.Kill(info.proc.UPID); err != nil {
		panic(err)
	}

	//run the action for failure
	info.a.Error("process timed out")
}
