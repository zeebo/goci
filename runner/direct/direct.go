package direct

import (
	gorpc "github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"fmt"
	"github.com/zeebo/goci/app/pinger"
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"log"
	"net/http"
	"os/exec"
	"runtime"
)

//Runner is an rpc service that runs tests locally.
type Runner struct {
	tcl    *client.Client  //the client for the tracker
	base   string          //the url the rpc server is hosted at
	rpc    *gorpc.Server   //the rpc server
	rq     rpc.RunnerQueue //the queue of run items
	runner string          //the path to the runner binary
	task   rpc.RunnerTask  //the current task being run
	resp   chan rpc.Output //the channel of responses

	key string //the key the tracker has stored us at
}

//New returns a new Runner ready to be Announced and run tests locally.
func New(runner, tracker, hosted string) *Runner {
	n := &Runner{
		tcl:    client.New(tracker, http.DefaultClient, client.JsonCodec),
		base:   hosted,
		runner: runner,
		rpc:    gorpc.NewServer(),
		rq:     rpc.NewRunnerQueue(),
		resp:   make(chan rpc.Output),
	}

	//register the run service in the rpc
	if err := n.rpc.RegisterService(n.rq, ""); err != nil {
		panic(err)
	}

	//register the pinger
	if err := n.rpc.RegisterService(pinger.Pinger{}, ""); err != nil {
		panic(err)
	}

	//register ourselves in the rpc
	if err := n.rpc.RegisterService(n, ""); err != nil {
		panic(err)
	}

	//register the codec
	n.rpc.RegisterCodec(json.NewCodec(), "application/json")

	//start processing
	go n.run()

	return n
}

//Announce tells the tracker that we're avaialable to run tests.
func (r *Runner) Announce() (err error) {
	args := &rpc.AnnounceArgs{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
		Type:   "Runner",
		URL:    r.base,
	}
	reply := new(rpc.AnnounceReply)
	if err = r.tcl.Call("Tracker.Announce", args, reply); err != nil {
		return
	}
	r.key = reply.Key
	return
}

//Remove removes this Runner from the tracker.
func (r *Runner) Remove() (err error) {
	args := &rpc.RemoveArgs{
		Key:  r.key,
		Kind: "Runner",
	}
	err = r.tcl.Call("Tracker.Remove", args, new(rpc.None))
	return
}

//ServeHTTP allows the runner to be hosted like any other http.Handler.
func (r *Runner) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.rpc.ServeHTTP(w, req)
}

//run grabs items from the queue and processes them.
func (r *Runner) run() {
	for {
		task := r.rq.Pop()
		r.process(task)
	}
}

//Post grabs the test output and sends it to the corresponding task managing it.
func (r *Runner) Post(req *http.Request, args *rpc.TestResponse, resp *rpc.None) (err error) {
	//make sure its for the id we're managing
	if args.ID != r.task.ID {
		err = rpc.Errorf("unknown ID: %s", args.ID)
		return
	}

	//send it the output
	r.resp <- args.Output
	return
}

//Request grabs the data for the test so the test runner can request the data
//it needs to run.
func (r *Runner) Request(req *http.Request, args *rpc.TestRequest, resp *rpc.RunTest) (err error) {
	//make sure its for the id we're managing
	if args.ID != r.task.ID {
		err = rpc.Errorf("unknown ID: %s", args.ID)
		return
	}

	//ensure the index is ok
	if args.Index < 0 || len(r.task.Tests) <= args.Index {
		err = rpc.Errorf("invalid index: %d not in [0, %d)", args.Index, len(r.task.Tests))
		return
	}

	//set the response
	*resp = r.task.Tests[args.Index]
	return
}

func (r *Runner) process(task rpc.RunnerTask) {
	log.Printf("Incoming task: %+v", task)

	//set the task and output slices up
	r.task = task
	outs := make([]rpc.Output, 0, len(task.Tests)+len(task.WontBuilds))
	outs = append(outs, task.WontBuilds...) //copy the wont builds in

	//start running all the tests
	for i := range task.Tests {
		cmd := exec.Command(r.runner, r.base, task.ID, fmt.Sprint(i))
		if err := cmd.Start(); err != nil {
			panic(err)
		}
	}

	//collect all the outputs
	for _ = range task.Tests {
		outs = append(outs, <-r.resp)
	}

	//build a runner response
	resp := &rpc.RunnerResponse{
		Key:      task.Key,
		ID:       task.ID,
		WorkRev:  task.WorkRev,
		Revision: task.Revision,
		RevDate:  task.RevDate,
		Tests:    outs,
	}

	log.Printf("Pushing response[%s]: %+v", task.Response, resp)

	//send it off
	cl := client.New(r.task.Response, http.DefaultClient, client.JsonCodec)
	if err := cl.Call("Response.Post", resp, new(rpc.None)); err != nil {
		log.Printf("Error pushing response: %s", err)
	}
}
