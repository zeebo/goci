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

//Work is the datastore entity representing the work item that came in
type Work struct {
	rpc.Work           //the parsed and distilled work item
	Data     string    //the raw data that came in
	Created  time.Time //when the item was received
}

//TaskInfo is an entity that stores when a task was sent out so the cron can
//readd tasks that have expired.
type TaskInfo struct {
	Key     string
	Created time.Time
}

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

func AddQueue(ctx appengine.Context, key string) (err error) {
	t := taskqueue.NewPOSTTask("/queue/work", url.Values{
		"key": {key},
	})
	_, err = taskqueue.Add(ctx, t, "work")
	return
}

func queueRequeue(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	//search the datastore for expired TaskInfo and throw the key back into the queue
	var vals []*TaskInfo
	keys, err := datastore.NewQuery("TaskInfo").
		Filter("Created < ", time.Now().Add(-5*time.Minute)).
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
		if err := AddQueue(ctx, info.Key); err != nil {
			e = httputil.Errorf(err, "error adding info to queue")
			return
		}
	}

	return
}

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
	err = cl.Call("Queue.Push", task, new(rpc.None))
	if err != nil {
		e = httputil.Errorf(err, "error pushing task into queue")
		return
	}

	return
}

//Response is a service that records Runner responses
type Response struct{}

func (Response) Post(req *http.Request, args *rpc.RunnerResponse, resp *rpc.None) (err error) {
	defer func() {
		//if we don't have an rpc.Error, encode it as one
		if _, ok := err.(rpc.Error); err != nil && !ok {
			err = rpc.Errorf("%s", err)
		}
	}()

	//create our context
	ctx := appengine.NewContext(req)
	ctx.Infof("Storing runner result")

	//make sure the TaskInfo for this request still exists, and if so, remove it
	var found bool
	trans := func(c appengine.Context) (err error) {
		info := new(TaskInfo)
		key := httputil.FromString(args.ID)
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

	//if we didn't find our TaskInfo, just bail
	if !found {
		ctx.Errorf("Got a late response")
		return
	}

	//store the response
	wkey := httputil.FromString(args.Key)
	key := datastore.NewIncompleteKey(ctx, "RunnerResponse", wkey)
	if _, err = datastore.Put(ctx, key, args); err != nil {
		return
	}

	//we did it!
	return
}
