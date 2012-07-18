// +build !goci

//package queue provides handlers for adding and dispatching work
package queue

import (
	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
	"appengine/urlfetch"
	gorcp "code.google.com/p/gorilla/rpc"
	gojson "code.google.com/p/gorilla/rpc/json"
	"httputil"
	"net/http"
	"net/url"
	"rpc"
	"rpc/client"
	"time"
	"tracker"
)

func init() {
	//create a new rpc server
	s := gorcp.NewServer()
	s.RegisterCodec(gojson.NewCodec(), "application/json")

	//add the response queue
	s.RegisterService(Response{}, "")

	//add our handlers including the rpc server
	http.Handle("/queue/work", httputil.Handler(queueWork))
	http.Handle("/queue/requeue", httputil.Handler(queueRequeue))
	http.Handle("/queue/response", s)
}

//Distiller is a type that can be added into the queue. It distills into a work
//item that can be sent to builders.
type Distiller interface {
	Distill() (rpc.Work, string)
}

//QueueWork takes a Distiller and adds it into the work queue.
func QueueWork(ctx appengine.Context, d Distiller) (err error) {
	//distill and create our work item
	work, data := d.Distill()
	q := &Work{
		Work:    work,
		Data:    data,
		Created: time.Now(),
	}

	//store it in the datastore
	key := datastore.NewIncompleteKey(ctx, "Work", nil)
	if key, err = datastore.Put(ctx, key, q); err != nil {
		return
	}

	//add the key to the queue
	if err = addQueue(ctx, httputil.ToString(key)); err != nil {
		return
	}

	return
}

//addQueue puts the key in the queue to be dispatched.
func addQueue(ctx appengine.Context, key string) (err error) {
	t := taskqueue.NewPOSTTask("/queue/work", url.Values{
		"key": {key},
	})
	_, err = taskqueue.Add(ctx, t, "work")
	return
}

//queueRequeue is a cron job that requeues all the tasks that took more than
//10 minutes to complete.
func queueRequeue(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	//search the datastore for expired TaskInfo and throw the key back into the queue
	var vals []*TaskInfo
	keys, err := datastore.NewQuery("TaskInfo").
		Filter("Created < ", time.Now().Add(-10*time.Minute)).
		GetAll(ctx, &vals)

	if err != nil {
		e = httputil.Errorf(err, "error getting keys to be purged")
		return
	}

	//bail early if we have no values
	if len(vals) == 0 {
		return
	}

	//delete them
	if err := datastore.DeleteMulti(ctx, keys); err != nil {
		e = httputil.Errorf(err, "error deleting keys to be purged")
	}

	//throw the keys back into the queue
	for _, info := range vals {
		if err := addQueue(ctx, info.Key); err != nil {
			e = httputil.Errorf(err, "error adding info to queue")
			return
		}
	}

	return
}

//queueWork is the handler that gets called for a queue item. It grabs a builder
//and runner and dispatches the work item to them, recoding when that operation
//started.
func queueWork(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	//grab the key of the work item
	key := httputil.FromString(req.FormValue("key"))

	//grab the work item
	work := new(Work)
	if err := datastore.Get(ctx, key, work); err != nil {
		e = httputil.Errorf(err, "error grabbing work item")
		return
	}

	//lease a builder and runner
	_, _, builder, runner, err := tracker.LeasePair(ctx)
	if err != nil {
		e = httputil.Errorf(err, "error leasing a pair of workers")
		return
	}

	//build the task info
	id := datastore.NewIncompleteKey(ctx, "TaskInfo", key)
	info := &TaskInfo{
		Key:     httputil.ToString(key),
		Created: time.Now(),
	}
	if id, err = datastore.Put(ctx, id, info); err != nil {
		e = httputil.Errorf(err, "error storing TaskInfo tracker")
		return
	}

	//build the task
	task := &rpc.BuilderTask{
		Work:     work.Work,
		Runner:   runner.URL,
		Response: "",
		Key:      httputil.ToString(key),
		ID:       httputil.ToString(id),
	}

	//send the task off to the builder queue
	cl := client.New(builder.URL, urlfetch.Client(ctx), client.JsonCodec)
	err = cl.Call("BuilderQueue.Push", task, new(rpc.None))
	if err != nil {
		e = httputil.Errorf(err, "error pushing task into queue")
		return
	}

	return
}

//Response is a service that records Runner responses
type Response struct{}

//findTaskInfo fetches and deletes the task info with the given id, returning
//if it was able to do so.
func findTaskInfo(ctx appengine.Context, id string) (found bool, err error) {
	trans := func(c appengine.Context) (err error) {
		info := new(TaskInfo)
		key := httputil.FromString(id)
		if err = datastore.Get(c, key, info); err != nil {
			return
		}
		if err = datastore.Delete(c, key); err != nil {
			return
		}
		found = true
		return
	}
	if err = datastore.RunInTransaction(ctx, trans, nil); err != nil {
		return
	}
	return
}

//Post is the rpc method that the Runner uses to give a response about an item.
func (Response) Post(req *http.Request, args *rpc.RunnerResponse, resp *rpc.None) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	//create our context
	ctx := appengine.NewContext(req)
	ctx.Infof("Storing runner result")

	//make sure the TaskInfo for this request still exists, and if so, remove it
	found, err := findTaskInfo(ctx, args.ID)

	//if we didn't find our TaskInfo, just bail
	if !found {
		ctx.Errorf("Got a late response")
		return
	}

	//TODO(zeebo): store the response

	//we did it!
	return
}

//Error is used when there were any errors in building the test
func (Response) Error(req *http.Request, args *rpc.BuilderResponse, resp *rpc.None) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	//create the context
	ctx := appengine.NewContext(req)
	ctx.Infof("Storing a builder error")

	//make sure the TaskInfo for this request still exists, and if so, remove it
	found, err := findTaskInfo(ctx, args.ID)

	//if we didn't find it, just bail
	if !found {
		ctx.Errorf("Got a late response")
		return
	}

	//TODO(zeebo): store the response

	//we did it!
	return
}
