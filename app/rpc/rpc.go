package rpc

import (
	"fmt"
	"time"
)

//Error is an error type suitable for sending over an rpc response
type Error string

//Error implemented the error interface for the Error type
func (r Error) Error() string {
	return string(r)
}

//Errorf is a helper method for creating an rpc error that can be transmitted
//in json.
func Errorf(format string, v ...interface{}) error {
	return Error(fmt.Sprintf(format, v...))
}

//Wrap is a convenience function to wrap whatever error we get in an rpc Error
func Wrap(errp *error) {
	if errp == nil {
		return
	}
	err := *errp
	if _, ok := err.(Error); err != nil && ok {
		*errp = Errorf("%s", err)
	}
}

//AnnounceArgs is the argument type of the Announce function
type AnnounceArgs struct {
	GOOS, GOARCH string //the goos/goarch of the service
	Type         string //either "Builder" or "Runner"
	URL          string //the url of the service to make rpc calls
}

//AnnounceReply is the reply type of the Announce function
type AnnounceReply struct {
	//Key is the datastore key that corresponds to the service if successful
	Key string
}

//RemoveArgs is the argument type of the Remove function
type RemoveArgs struct {
	//Key is the datastore key that corresponds to the service to be removed
	Key string
}

//None is an empty rpc element
type None struct{}

//Work is an incoming work item to generate the builds for a given revision and
//import path. If Revision is empty, the revision chosen by go get is used. If
//Subpackages is true, it will build binaries for all subpackages of the import
//path as well.
type Work struct {
	Revision    string
	ImportPath  string
	Subpackages bool

	//VCSHint is an optional parameter that specifies the version control system
	//used by the package. If set to the empty string, we will search for the 
	//system by looking for the metadata directory.
	VCSHint string
}

//Distill makes a Work able to be sent in to the queue.
func (w Work) Distill() (Work, string) {
	return w, ""
}

//BuilderTask is a task sent to a Builder
type BuilderTask struct {
	Work     Work   //the Work item to be completed
	Key      string //the datastore key for the Work item (forward to runner)
	ID       string //the ID of the TaskInfo in the datastore (forward to runner)
	Runner   string //the rpc url of the runner for this task
	Response string //the rpc url of the response (forward to the runner)
}

//RunnerTask is a task sent by a Builder to a runner
type RunnerTask struct {
	Key      string    //the key of the work item
	ID       string    //the id of the task
	Response string    //the rpc url of the response
	Revision string    //the revision we ended up testing
	RevDate  time.Time //the time this revision was made
	Tasks    []RunTest //the set of binarys to be executed
}

//RunTask represents an individual binary to be installed and run.
type RunTest struct {
	BinaryURL  string //the url to download the binary
	SourceURL  string //the url to download the tarball
	ImportPath string //the import path of the packge the binary is testing
}

//RunnerResponse is the response from the Runner to the tracker
type RunnerResponse struct {
	Key         string    //the datastore key for the Work item
	ID          string    //the ID of the TaskInfo in the datastore
	BuildOutput string    //the output of the build phase
	Revision    string    //the revision we ended up testing
	RevDate     time.Time //the time this revision was made
	Outputs     []Output  //the list of outputs from the tests
}

//BuilderResponse is the response from the Builder if the build failed for any
//reason.
type BuilderResponse struct {
	Key         string   //the key of the work item
	ID          string   //the id of the task
	Error       string   //the error in setting up the builds
	BuildErrors []Output //the errors in building the binaries
}

//Output is a type that wraps the output of a build, be it the actual output or
//the error produced.
type Output struct {
	ImportPath string //the import path of the binary that produced the output
	Output     string //the output of the test
}
