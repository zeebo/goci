package web

import (
	gorpc "code.google.com/p/gorilla/rpc"
	"code.google.com/p/gorilla/rpc/json"
	"github.com/zeebo/goci/app/pinger"
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"github.com/zeebo/goci/heroku"
	"net/http"
	"time"
)

type Runner struct {
	app, api string
	tcl      *client.Client
	base     string
	rpc      *gorpc.Server
	rq       rpc.RunnerQueue
	mc       *heroku.ManagedClient
	tm       *runnerTaskMap

	key string
}

func New(app, api string, tracker, hosted string) *Runner {
	n := &Runner{
		app:  app,
		api:  api,
		tcl:  client.New(tracker, http.DefaultClient, client.JsonCodec),
		base: hosted,
		rpc:  gorpc.NewServer(),
		rq:   rpc.NewRunnerQueue(),
		mc:   heroku.NewManaged(app, api, 2, 2*time.Minute),
		tm:   &runnerTaskMap{items: map[string]*runnerTask{}},
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

//Announce tells the tracker that we're available to run tests.
func (r *Runner) Announce() (err error) {
	args := &rpc.AnnounceArgs{
		GOOS:   "linux",
		GOARCH: "amd64",
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

//Remove removes this Builder from the tracker.
func (r *Runner) Remove() (err error) {
	args := &rpc.RemoveArgs{Key: r.key}
	err = r.tcl.Call("Tracker.Remove", args, new(rpc.None))
	return
}

//ServeHTTP allows the builder to be hosted like any other http.Handler.
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
