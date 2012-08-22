package workqueue

import (
	"github.com/zeebo/goci/app/rpc"
	"labix.org/v2/mgo/bson"
	"time"
)

//Work is the datastore entity representing the work item that came in
type Work struct {
	ID bson.ObjectId `bson:"_id,omitempty"`

	Work       rpc.Work  //the parsed and distilled work item
	Data       string    //the raw data that came in
	Created    time.Time //when the item was received
	Status     string    //the current status of the work item
	AttemptLog []Attempt //the attempts building the work item

	Lock lock //lock for the document
}

//Attempt represents an attempt to build a Work item
type Attempt struct {
	When    time.Time     //when the attempt was started
	Builder string        //the builder the attempt was for
	Runner  string        //the runner the attempt was for
	ID      bson.ObjectId //a random ID for the attempt
}

//lock represents a lock on a document
type lock struct {
	Expires time.Time //when the lock expires
	Who     string    //who owns the lock
}

//define some string constants for statuses
const (
	statusWaiting    = "waiting"
	statusProcessing = "processing"
	statusCompleted  = "completed"
)
