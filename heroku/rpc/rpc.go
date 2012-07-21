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
