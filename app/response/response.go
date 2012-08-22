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

	//TODO(zeebo): search for the TaskInfo from the queue
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

	//TODO(zeebo): search for the task info from the queue

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

//DispatchError is called when there were too many errors attempting to dispatch
//the work item and the dispatcher gave up trying to send it out. This could
//happen if there is a misbehabving builder or runner, or if the test takes
//too long to run and the response window closes.
func (Response) DispatchError(req *http.Request, args *rpc.DispatchResponse, resp *rpc.None) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	//create the context
	ctx := httputil.NewContext(req)

	//get the key of the work item
	key := bson.ObjectIdHex(args.Key)

	//create a WorkResult
	w := &WorkResult{
		ID:      bson.NewObjectId(),
		WorkID:  key,
		Success: false,
		When:    time.Now(),
		Error:   []byte(args.Error),
	}

	//store it
	if err = ctx.DB.C("WorkResult").Insert(w); err != nil {
		ctx.Errorf("Error storing WorkResult: %v", err)
		return
	}

	//we did it!
	return
}
