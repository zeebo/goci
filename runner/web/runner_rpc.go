package web

import (
	"github.com/zeebo/goci/app/rpc"
	"net/http"
)

//Post grabs the test output and sends it to the corresponding task managing
//it.
func (r *Runner) Post(req *http.Request, args *rpc.TestResponse, resp *rpc.None) (err error) {
	//grab the task managing this output
	task, ok := r.tm.Lookup(args.ID)
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
func (r *Runner) Request(req *http.Request, args *rpc.TestRequest, resp *rpc.RunTest) (err error) {
	//grab the task managing this request
	task, ok := r.tm.Lookup(args.ID)
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
