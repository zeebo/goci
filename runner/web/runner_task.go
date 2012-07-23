package web

import (
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"github.com/zeebo/goci/heroku"
	"log"
	"net/http"
	"sync"
)

//runerTaskMap stores a mapping of ids to runnerTasks
type runnerTaskMap struct {
	sync.Mutex
	items map[string]*runnerTask
}

//Register stores the path for later retieval.
func (m *runnerTaskMap) Register(task *runnerTask) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.items[task.task.ID]; ok {
		panic("id already exists")
	}

	//store it
	m.items[task.task.ID] = task
	return
}

//Lookup gets the path for the given id.
func (m *runnerTaskMap) Lookup(id string) (task *runnerTask, ok bool) {
	m.Lock()
	defer m.Unlock()

	task, ok = m.items[id]
	return
}

//Delete removes the id from the map.
func (m *runnerTaskMap) Delete(id string) {
	m.Lock()
	defer m.Unlock()

	delete(m.items, id)
}

//runnerTask represents a runner task in progress.
type runnerTask struct {
	mc *heroku.ManagedClient
	tm *runnerTaskMap

	task  rpc.RunnerTask
	resps chan rpc.Output
	ids   map[string]chan string
}

//run grabs all the items from the channel and sends in a response
func (r *runnerTask) run() {
	//grab all of the output
	outs := make([]rpc.Output, 0, cap(r.resps))
	for i := 0; i < cap(r.resps); i++ {
		//grab an outout
		o := <-r.resps

		//signal to the default client that we're finished with this id
		idch := r.ids[o.ImportPath]
		r.mc.Finished(<-idch)

		//append the output
		outs = append(outs, o)
	}

	//we're done grabbing output so delete ourselves from the task map
	r.tm.Delete(r.task.ID)

	//build a RunnerResponse
	resp := &rpc.RunnerResponse{
		Key:      r.task.Key,
		ID:       r.task.ID,
		Revision: r.task.Revision,
		RevDate:  r.task.RevDate,
		Outputs:  outs,
	}

	log.Printf("Pushing response[%s]: %+v", r.task.Response, resp)

	//send if off
	cl := client.New(r.task.Response, http.DefaultClient, client.JsonCodec)
	if err := cl.Call("Response.Post", resp, new(rpc.None)); err != nil {

	}
}
