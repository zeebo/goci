package rpc

import (
	"fmt"
	"github.com/zeebo/goci/builder"
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

//BuilderTask is a task sent to a Builder
type BuilderTask struct {
	Work   builder.Work //the Work item to be completed
	Runner string       //the rpc url of the runner for this task
}
