//package rpc defines some structs that the runner process and the main heroku
//process use to communicate.
package rpc

import (
	"github.com/zeebo/goci/app/rpc"
)

type TestResponse struct {
	Output rpc.Output //the output of the test 
	ID     string     //the ID for the test
}

type TestRequest struct {
	ID    string //the ID of the test
	Index int    //the index of the test to be run
}
