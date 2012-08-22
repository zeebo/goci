package response

import (
	"labix.org/v2/mgo/bson"
	"time"
)

//TestResult is an entity type that describes the sucessful running of a test
//and the output it generated.
type TestResult struct {
	ID           bson.ObjectId `bson:"_id,omitempty"`
	WorkResultID bson.ObjectId //key of the work result that this came from

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
