//package entities contains all the database entities
package entities

import (
	"github.com/zeebo/goci/app/rpc"
	"labix.org/v2/mgo/bson"
	"time"
)

type Notification struct {
	ID bson.ObjectId `bson:"_id,omitempty"`

	Test       bson.ObjectId  //the id of the test
	Config     rpc.Config     //the configuration of the test
	Status     string         //the status of the notification
	AttemptLog []NotifAttempt //the attempts sending out the notification
	Revision   int            //the revision of the notification doc
}

//NotifAttempt represents an attempt to send out notifications
type NotifAttempt struct {
	ID   bson.ObjectId //a random id for the test attempt
	When time.Time     //when the attempt was started
}

//define some string constants for statuses
const (
	NotifStatusError      = "error"
	NotifStatusWaiting    = "waiting"
	NotifStatusProcessing = "processing"
	NotifStatusCompleted  = "completed"
)

//TestResult is an entity type that describes the sucessful running of a test
//and the output it generated.
type TestResult struct {
	ID           bson.ObjectId `bson:"_id,omitempty" json:"-"`
	WorkResultID bson.ObjectId `json:"-"` //key of the work result that this came from

	ImportPath string    //import path of the test
	Revision   string    //revision of the source code
	RevDate    time.Time //when the revision was commit
	When       time.Time //when the test result was recorded
	Output     string    //the output of the test
	Status     string    //the status of the build: (Pass/Fail/WontBuild/Error)
}

//WorkResult is an entity type that represents the result of the work item being
//run through the queue. It records any build failures or other errors in
//generating the test results.
type WorkResult struct {
	ID     bson.ObjectId `bson:"_id,omitempty"`
	WorkID bson.ObjectId //id of the work item that generated this result

	Success  bool      //all the work items completed sucessfully
	Revision string    //the revision of the build (if known)
	RevDate  time.Time //when the revision was commit (if known)
	When     time.Time //the time the work response was recorded
	Error    string    //a general error before any code could be built
}

//Work is the datastore entity representing the work item that came in
type Work struct {
	ID bson.ObjectId `bson:"_id,omitempty"`

	Work       rpc.Work      //the parsed and distilled work item
	Data       string        //the raw data that came in
	Created    time.Time     //when the item was received
	Status     string        //the current status of the work item
	AttemptLog []WorkAttempt //the attempts building the work item
	Revision   int           //the revision number of the work item
}

//WorkAttempt represents an attempt to build a Work item
type WorkAttempt struct {
	ID      bson.ObjectId //a random id for the test attempt
	When    time.Time     //when the attempt was started
	Builder string        //the builder the attempt was for
	Runner  string        //the runner the attempt was for
}

//define some string constants for statuses
const (
	WorkStatusWaiting    = "waiting"
	WorkStatusProcessing = "processing"
	WorkStatusCompleted  = "completed"
)
