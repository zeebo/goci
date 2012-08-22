//package workqueue provides handlers for adding and dispatching work
package workqueue

import (
	"fmt"
	"github.com/zeebo/goci/app/httputil"
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"github.com/zeebo/goci/app/tracker"
	"labix.org/v2/mgo/bson"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	http.Handle("/queue/dispatch", httputil.Handler(dispatchWork))
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
	wkey := bson.NewObjectId()
	q := &Work{
		ID:      wkey,
		Work:    work,
		Data:    data,
		Status:  statusWaiting,
		Created: time.Now(),
	}

	//store it in the datastore
	if err = ctx.DB.C("Work").Insert(q); err != nil {
		return
	}

	//send a request to dispatch the queue
	go http.Get(httputil.Absolute("/queue/dispatch"))

	return
}

var (
	lockTime    = 30 * time.Second
	attemptTime = 10 * time.Minute
	maxAttempts = 5
)

//dispatchWork is the handler that gets called for a queue item. It grabs a builder
//and runner and dispatches the work item to them, recoding when that operation
//started.
func dispatchWork(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	//generate a unique id for this request
	name := fmt.Sprint(rand.Int63())

	//selector is a bson document corresponding to selecting unlocked work items
	//that are either waiting, or processing and their most recent attempt is
	//over attemptTime old.
	type L []interface{}
	selector := bson.M{
		"$and": L{
			bson.M{"lock.expires": bson.M{"$lt": time.Now()}},
			bson.M{"$or": L{
				bson.M{"status": statusWaiting},
				bson.M{"$and": L{
					bson.M{"status": statusProcessing},
					bson.M{"attemptlog.0.when": bson.M{"$lt": time.Now().Add(-1 * attemptTime)}},
				}},
			}},
		},
	}
	//update acquires the lock for the work item by setting it to now, leasing
	//the work item to this function for lockTime.
	update := bson.D{
		{"$set", bson.M{
			"lock.expires": time.Now().Add(lockTime),
			"lock.who":     name,
		}},
	}

	//run the query locking documents we're going to use
	info, err := ctx.DB.C("Work").UpdateAll(selector, update)
	if err != nil {
		e = httputil.Errorf(err, "error grabbing work items from the queue")
		return
	}

	//bail early if we locked no documents
	if info.Updated == 0 {
		ctx.Infof("Locked no Work items. Exiting")
		return
	}

	//attempt to unlock all of the documents when the function exits
	defer ctx.DB.C("Work").UpdateAll(
		bson.M{"lock.who": name}, //selector
		bson.M{"$set": bson.M{ //update 
			"lock.expires": time.Time{},
		}},
	)

	//loop over our locked documents and dispatch work to them
	iter := ctx.DB.C("Work").Find(bson.M{"lock.who": name}).Iter()

	var work Work
	for iter.Next(&work) {
		//if the status is waiting or the attemptlog is short enough dispatch
		//it to a builder
		if work.Status == statusWaiting || len(work.AttemptLog) < maxAttempts {
			if err := dispatchWorkItem(ctx, work); err != nil {
				ctx.Infof("Error dispatching work item: %s", err)
			}
			continue
		}

		//we have a work item with too many attempts, so let's flag it completed
		//and fire off an rpc call to store it as failed.
		ctx.Infof("Work item %s had too many attempts", work.ID)
		resp := &rpc.DispatchResponse{
			Key:   work.ID.Hex(),
			Error: "Unable to complete Work item after 5 attempts",
		}
		cl := client.New(httputil.Absolute("/rpc/response"), http.DefaultClient, client.JsonCodec)
		if err := cl.Call("Response.DispatchError", resp, new(rpc.None)); err != nil {
			ctx.Infof("Couldn't store a dispatch error for work item %s: %s", work.ID, err)
			continue
		}

		//now flag the work item as completed in the queue
		err := ctx.DB.C("Work").Update(bson.M{"_id": work.ID}, bson.M{
			"$set": bson.M{"status": statusCompleted},
		})

		//log if we had an error marking it as completed, but this isn't fatal
		if err != nil {
			ctx.Infof("Error setting status after dispatching error for work item %s: %s", work.ID, err)
		}
	}

	//check for errors running the iteration
	if err = iter.Err(); err != nil {
		ctx.Errorf("Error iterating over work items: %s", err)
		e = httputil.Errorf(err, "Error iterating over work items")
		return
	}

	return
}

func dispatchWorkItem(ctx httputil.Context, work Work) (err error) {
	//lease a builder and runner
	builder, runner, err := tracker.LeasePair(ctx)
	if err != nil {
		return
	}

	log.Printf("Got:\nBuilder: %+v\nRunner: %+v", builder, runner)

	//create an attempt
	a := Attempt{
		When:    time.Now(),
		Builder: builder.URL,
		Runner:  runner.URL,
		ID:      bson.NewObjectId(),
	}

	//build the task
	task := &rpc.BuilderTask{
		Work:     work.Work,
		Runner:   runner.URL,
		Response: httputil.Absolute("/rpc/response"),
		Key:      work.ID.Hex(),
		ID:       a.ID.Hex(),
	}

	//be sure to send off the task first because even if we fail to store the
	//attempt in the database, we'll just do some extra work. If we stored and
	//send the task first, we could have an attempt that had no action, which
	//is worse.

	//send the task off to the builder queue
	cl := client.New(builder.URL, http.DefaultClient, client.JsonCodec)
	err = cl.Call("BuilderQueue.Push", task, new(rpc.None))
	if err != nil {
		return
	}

	//push the new attempt at the start of the array
	log := append([]Attempt{a}, work.AttemptLog...)

	//store the attempt on the document
	err = ctx.DB.C("Work").Update(bson.M{"_id": work.ID}, bson.D{
		{"$set", bson.M{"status": statusProcessing}},
		{"$set", bson.M{"attemptlog": log}},
	})
	return
}
