package main

import (
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	hrpc "github.com/zeebo/goci/heroku/rpc"
	"log"
	"net/http"
	"sync"
)

//defaultTaskMap is the default map of ID to open tasks
var defaultTaskMap = &runnerTaskMap{items: map[string]*runnerTask{}}

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
		bail("id already exists")
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
		defaultClient.Finished(<-idch)

		//append the output
		outs = append(outs, o)
	}

	//we're done grabbing output so delete ourselves from the task map
	defaultTaskMap.Delete(r.task.ID)

	//build a RunnerResponse
	resp := &rpc.RunnerResponse{
		Key:      r.task.Key,
		ID:       r.task.ID,
		Revision: r.task.Revision,
		RevDate:  r.task.RevDate,
		Outputs:  outs,
	}

	log.Printf("Sending response: %+v", resp)

	//send if off
	cl := client.New(r.task.Response, http.DefaultClient, client.JsonCodec)
	if err := cl.Call("Response.Post", resp, new(rpc.None)); err != nil {
		log.Println("Error sending runner response:", err)
	}
}

//RunManager is an rpc service for handling requests and responses from the
//process running the test.
type RunManager struct{}

//Post grabs the test output and sends it to the corresponding task managing
//it.
func (RunManager) Post(req *http.Request, args *hrpc.TestResponse, resp *rpc.None) (err error) {
	//grab the task managing this output
	task, ok := defaultTaskMap.Lookup(args.ID)
	if !ok {
		err = rpc.Errorf("unknown ID: %s", args.ID)
		return
	}

	//send it the output
	task.resps <- args.Output
	return
}

//Request grabs the data for the test so the test runner can request the data
//it needs to run.
func (RunManager) Request(req *http.Request, args *hrpc.TestRequest, resp *rpc.RunTest) (err error) {
	//grab the task managing this request
	task, ok := defaultTaskMap.Lookup(args.ID)
	if !ok {
		err = rpc.Errorf("unknown ID: %s", args.ID)
		return
	}

	//ensure the index is ok
	if args.Index < 0 || len(task.task.Tests) <= args.Index {
		err = rpc.Errorf("invalid index: %d not in [0, %d)", args.Index, len(task.task.Tests))
		return
	}

	//set the response
	*resp = task.task.Tests[args.Index]
	return
}

//register our RunManager
func init() {
	if err := rpcServer.RegisterService(RunManager{}, ""); err != nil {
		bail(err)
	}
}
