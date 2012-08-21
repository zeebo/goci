// +build !goci

//package response provides handlers for storing responses
package response

import (
	gorcp "code.google.com/p/gorilla/rpc"
	gojson "code.google.com/p/gorilla/rpc/json"
	"fmt"
	"github.com/zeebo/goci/app/httputil"
	"net/http"
	// "queue"
	"github.com/zeebo/goci/app/rpc"
	"labix.org/v2/mgo/bson"
	"strings"
	"time"
)

func init() {
	//create a new rpc server
	s := gorcp.NewServer()
	s.RegisterCodec(gojson.NewCodec(), "application/json")

	//add the response queue
	s.RegisterService(Response{}, "")

	http.Handle("/rpc/response", s)
}

//Response is a service that records Runner responses
type Response struct{}

//Post is the rpc method that the Runner uses to give a response about an item.
func (Response) Post(req *http.Request, args *rpc.RunnerResponse, resp *rpc.None) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	//create our context
	ctx := httputil.NewContext(req)
	ctx.Infof("Storing runner result")
	ctx.Infof("%+v", args)

	//TODO(zeebo): search for the TaskInfo from the queue
	/*
		//make sure the TaskInfo for this request still exists, and if so, remove it
		found, err := queue.FindTaskInfo(ctx, args.ID)

		if err != nil {
			ctx.Errorf("error finding task info: %v", err)
			return
		}

		//if we didn't find our TaskInfo, just bail
		if !found {
			ctx.Errorf("Got a late response")
			return
		}
	*/

	//TODO(zeebo): put this in a transaction to skip partial test results

	//get the key of the work item
	key := bson.ObjectIdHex(args.Key)

	//create a WorkResult
	wkey := bson.NewObjectId()
	w := &WorkResult{
		ID:       wkey,
		WorkID:   key,
		Success:  true,
		Revision: args.Revision,
		RevDate:  args.RevDate,
		When:     time.Now(),
	}

	//store it
	if err = ctx.DB.C("WorkResult").Insert(w); err != nil {
		ctx.Errorf("Error storing WorkResult: %v", err)
		return
	}

	//store the test results
	for _, out := range args.Tests {

		//get the status from the output type and output
		var status string
		switch out.Type {
		case rpc.OutputSuccess:
			if strings.HasSuffix(out.Output, "\nPASS\n") {
				status = "Pass"
			} else {
				status = "Fail"
			}
		case rpc.OutputWontBuild:
			status = "WontBuild"
		case rpc.OutputError:
			status = "Error"
		default:
			err = fmt.Errorf("unknown output type: %s", out.Type)
			return
		}

		//make a TestResult
		t := &TestResult{
			ID:           bson.NewObjectId(),
			WorkResultID: wkey,
			ImportPath:   out.ImportPath,
			Revision:     args.Revision,
			RevDate:      args.RevDate,
			When:         time.Now(),
			Output:       []byte(out.Output),
			Status:       status,
		}

		//store it
		if err = ctx.DB.C("TestResult").Insert(t); err != nil {
			ctx.Errorf("Error storing TestResult: %v", err)
			return
		}
	}

	return
}

//Error is used when there were any errors in building the test
func (Response) Error(req *http.Request, args *rpc.BuilderResponse, resp *rpc.None) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	//create the context
	ctx := httputil.NewContext(req)
	ctx.Infof("Storing a builder error")
	ctx.Infof("%+v", args)

	//TODO(zeebo): search for the task info from the queue
	/*
		//make sure the TaskInfo for this request still exists, and if so, remove it
		found, err := queue.FindTaskInfo(ctx, args.ID)

		//make sure we check the error
		if err != nil {
			ctx.Errorf("error finding task info: %v", err)
			return
		}

		//if we didn't find it, just bail
		if !found {
			ctx.Errorf("Got a late response")
			return
		}
	*/

	//get the key of the work item
	key := bson.ObjectIdHex(args.Key)

	//create a WorkResult
	w := &WorkResult{
		ID:       bson.NewObjectId(),
		WorkID:   key,
		Success:  false,
		When:     time.Now(),
		Revision: args.Revision,
		RevDate:  args.RevDate,
		Error:    []byte(args.Error),
	}

	//store it
	if err = ctx.DB.C("WorkResult").Insert(w); err != nil {
		ctx.Errorf("Error storing WorkResult: %v", err)
		return
	}

	//we did it!
	return
}
