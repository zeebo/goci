//package workqueue provides handlers for adding and dispatching work
package workqueue

import (
	"github.com/zeebo/goci/app/entities"
	"github.com/zeebo/goci/app/httputil"
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"github.com/zeebo/goci/app/rpc/router"
	"github.com/zeebo/goci/app/tracker"
	"labix.org/v2/mgo/bson"
	"labix.org/v2/mgo/txn"
	"log"
	"net/http"
	"time"
)

const handleUrl = "/queue/dispatch"

func init() {
	http.Handle(handleUrl, httputil.Handler(dispatchWork))
}

//Distiller is a type that can be added into the queue. It distills into a work
//item that can be sent to builders. The second item is a string representation
//of the raw data that is being distilled.
type Distiller interface {
	Distill() (rpc.Work, string)
}

//QueueWork takes a Distiller and adds it into the work queue.
func QueueWork(ctx httputil.Context, d Distiller) (err error) {
	//distill and create our work item
	work, data := d.Distill()
	q := &entities.Work{
		ID:      bson.NewObjectId(),
		Work:    work,
		Data:    data,
		Status:  entities.WorkStatusWaiting,
		Created: time.Now(),
	}

	//store it in the datastore
	if err = ctx.DB.C("Work").Insert(q); err != nil {
		return
	}

	//send a request to dispatch the queue
	go http.Get(httputil.Absolute(handleUrl))

	return
}

const (
	attemptTime = 10 * time.Minute
	maxAttempts = 5
)

//dispatchWork is the handler that gets called for a queue item. It grabs a builder
//and runner and dispatches the work item to them, recoding when that operation
//started.
func dispatchWork(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	//find all the documents that are waiting or (processing and their attempt is
	//taking too long)
	type L []interface{}
	selector := bson.M{
		"$or": L{
			bson.M{"status": entities.WorkStatusWaiting},
			bson.M{
				"status":            entities.WorkStatusProcessing,
				"attemptlog.0.when": bson.M{"$lt": time.Now().Add(-1 * attemptTime)},
			},
		},
	}
	iter := ctx.DB.C("Work").Find(selector).Iter()

	var work entities.Work
	for iter.Next(&work) {
		//if its a processing task with too many attempts, store it as a dispatch
		//error.
		if len(work.AttemptLog) >= maxAttempts {
			ctx.Infof("Work item %s had too many attempts", work.ID)
			args := &rpc.DispatchResponse{
				Key:     work.ID.Hex(),
				Error:   "Unable to complete Work item. Too many failed attempts.",
				WorkRev: work.Revision,
			}

			//send it off to the response rpc
			respUrl := httputil.Absolute(router.Lookup("Response"))
			cl := client.New(respUrl, http.DefaultClient, client.JsonCodec)
			if err := cl.Call("Response.DispatchError", args, new(rpc.None)); err != nil {
				ctx.Infof("Couldn't store a dispatch error for work item %s: %s", work.ID, err)
			}

			continue
		}

		//attempt to dispatch the work item
		err := dispatchWorkItem(ctx, work)
		if err != nil {
			ctx.Errorf("Error dispatching work: %s", err)
		}
	}

	//check for errors running the iteration
	if err := iter.Err(); err != nil {
		ctx.Errorf("Error iterating over work items: %s", err)
		e = httputil.Errorf(err, "Error iterating over work items")
		return
	}

	return
}

func dispatchWorkItem(ctx httputil.Context, work entities.Work) (err error) {
	//lease a builder and runner
	builder, runner, err := tracker.LeasePair(ctx)
	if err != nil {
		return
	}

	log.Printf("Got:\nBuilder: %+v\nRunner: %+v", builder, runner)

	//create an attempt
	a := entities.WorkAttempt{
		When:    time.Now(),
		Builder: builder.URL,
		Runner:  runner.URL,
		ID:      bson.NewObjectId(),
	}

	//push the new attempt at the start of the array
	log := append([]entities.WorkAttempt{a}, work.AttemptLog...)

	//transactionally acquire ownership of the document
	ops := []txn.Op{{
		C:  "Work",
		Id: work.ID,
		Assert: bson.M{
			"revision": work.Revision,
		},
		Update: bson.M{
			"$inc": bson.M{"revision": 1},
			"$set": bson.M{
				"attemptlog": log,
				"status":     entities.WorkStatusProcessing,
			},
		},
	}}

	err = ctx.R.Run(ops, bson.NewObjectId(), nil)
	if err == txn.ErrAborted {
		ctx.Infof("Lost the race dispatching a work item")
		err = nil
		return
	}
	if err != nil {
		return
	}

	//build the task
	task := &rpc.BuilderTask{
		Work:     work.Work,
		Key:      work.ID.Hex(),
		ID:       a.ID.Hex(),
		WorkRev:  work.Revision + 1,
		Runner:   runner.URL,
		Response: httputil.Absolute(router.Lookup("Response")),
	}

	//send the task off to the builder queue
	cl := client.New(builder.URL, http.DefaultClient, client.JsonCodec)
	err = cl.Call("BuilderQueue.Push", task, new(rpc.None))

	return
}
