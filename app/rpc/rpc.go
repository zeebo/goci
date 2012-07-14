package rpc

import "fmt"

//Error is an error type suitable for sending over an rpc response
type Error string

func (r Error) Error() string {
	return string(r)
}

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
	Key string
}

//RemoveReply is the reply type of the Remove function
type RemoveReply struct{}
