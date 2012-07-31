// +build !goci

//package response provides handlers for storing responses
package response

import (
	"appengine"
	gorcp "code.google.com/p/gorilla/rpc"
	gojson "code.google.com/p/gorilla/rpc/json"
	"net/http"
	"queue"
	"rpc"
)

func init() {
	//create a new rpc server
	s := gorcp.NewServer()
	s.RegisterCodec(gojson.NewCodec(), "application/json")

	//add the response queue
	s.RegisterService(Response{}, "")

	http.Handle("/response", s)
}

//Response is a service that records Runner responses
type Response struct{}

//Post is the rpc method that the Runner uses to give a response about an item.
func (Response) Post(req *http.Request, args *rpc.RunnerResponse, resp *rpc.None) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	//create our context
	ctx := appengine.NewContext(req)
	ctx.Infof("Storing runner result")
	ctx.Infof("%+v", args)

	//make sure the TaskInfo for this request still exists, and if so, remove it
	found, err := queue.FindTaskInfo(ctx, args.ID)

	//if we didn't find our TaskInfo, just bail
	if !found {
		ctx.Errorf("Got a late response")
		return
	}

	//TODO(zeebo): store the response

	//we did it!
	return
}

//Error is used when there were any errors in building the test
func (Response) Error(req *http.Request, args *rpc.BuilderResponse, resp *rpc.None) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	//create the context
	ctx := appengine.NewContext(req)
	ctx.Infof("Storing a builder error")
	ctx.Infof("%+v", args)

	//make sure the TaskInfo for this request still exists, and if so, remove it
	found, err := queue.FindTaskInfo(ctx, args.ID)

	//if we didn't find it, just bail
	if !found {
		ctx.Errorf("Got a late response")
		return
	}

	//TODO(zeebo): store the response

	//we did it!
	return
}
