package rpc

import "fmt"

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

//BuilderTask is a task sent to a Builder
type BuilderTask struct {
	Work     Work   //the Work item to be completed
	Key      string //the datastore key for the Work item (forward to runner)
	ID       string //the ID of the TaskInfo in the datastore (forward to runner)
	Runner   string //the rpc url of the runner for this task
	Response string //the rpc url of the response (forward to the runner)
}

//RunnerResponse is the response from the Runner to the tracker
type RunnerResponse struct {
	Key         string //the datastore key for the Work item
	ID          string //the ID of the TaskInfo in the datastore
	BuildOutput string //the output of the build phase
	Outputs     []struct {
		ImportPath string //the import path of the binary
		Output     string //the output from running it
	}
}
