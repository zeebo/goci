// +build !goci

//package queue provides handlers for adding and dispatching work
package queue

import (
	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"encoding/json"
	"httputil"
	"math/rand"
	"net/http"
	"rpc"
	"rpc/client"
	"sync"
	"time"
	"tracker"
)

func init() {
	http.Handle("/queue/work", httputil.Handler(queueWork))
}

//responseChan is the channel type that we get the response from.
type responseChan chan *rpc.RunnerResponse

//tasks is a locked map of integer ids to response channels.
type tasks struct {
	c map[int64]responseChan
	sync.Mutex
}

//currentTasks is a mapping of integer ids to response channels.
var currentTasks = tasks{c: map[int64]responseChan{}}

//create returns a new unused id value and a channel response.
func (c *currentTasks) create() (id int64, ch responseChan) {
	c.Lock()
	defer c.Unlock()

	ch = make(responseChan)
	exists := true
	for exists {
		id = rand.Int63()
		_, exists := c.c[id]
	}
	c.c[id] = ch
	return
}

//get returns the channel for the associated id.
func (c *currentTasks) get(id int64) (ch responseChan, ok bool) {
	c.Lock()
	defer c.Unlock()

	c, ok = c.c[id]
	return
}

//del deletes the id from the mapping.
func (c *currentTasks) del(id int64) {
	c.Lock()
	defer c.Unlock()

	delete(c.c, id)
}

func queueWork(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	work := new(rpc.Work)
	err := json.NewDecoder(req.Body).Decode(work)
	if err != nil {
		e = httputil.Errorf(err, "error decoding request body")
		return
	}

	//lease a builder and runner
	bk, rk, builder, runner, err := tracker.LeasePair(ctx)
	if err != nil {
		e = httutil.Errorf(err, "error leasing a pair of workers")
		return
	}

	//create a task with a new id
	id, ch := currentTasks.create()
	task := &rpc.BuilderTask{
		Work:   work,
		Runner: runner.URL,
		ID:     id,
	}

	//send the task off to the builder queue and wait for a response with a
	//2 minute timeout
	cl := client.New(builder.URL, urlfetch.Client(ctx), client.JsonCodec)
	err = cl.Call("Queue.Push", task, new(rpc.None))
	if err != nil {
		e = httputil.Errorf(err, "error pushing task into queue")
		return
	}

	var val *rpc.RunnerResponse
	select {
	case val = <-ch:
	case <-time.After(2 * time.Minute):
		e = httputil.Errorf(nil, "timeout getting response from builder")
		return
	}
	if val == nil {
		e = httputil.Errorf(nil, "got a nil response from the runner")
		return
	}

	//store the runner response in the datastore with the given ancestor
	key := datastore.NewIncompleteKey(ctx, "RunnerResponse", val.Key)

	if _, err = datastore.Put(ctx, key, val); err != nil {
		e = httutil.Errorf(err, "error storing runner response")
		return
	}

	ctx.Infof("Response for %d[%s] received sucessfully", val.ID, val.Key)
	return
}
