//package response provides handlers for storing responses
package response

import (
	gorcp "code.google.com/p/gorilla/rpc"
	gojson "code.google.com/p/gorilla/rpc/json"
	"fmt"
	"github.com/zeebo/goci/app/httputil"
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/workqueue"
	"labix.org/v2/mgo/bson"
	"labix.org/v2/mgo/txn"
	"net/http"
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
	defer ctx.Close()

	//build the keys we need to reference
	key := bson.ObjectIdHex(args.key)
	wkey := bson.NewObjectId()

	ops := []txn.Operation{{ //make sure we have the given work item
		Collection: "Work",
		DocId:      key,
		Assert: bson.M{
			"status":          workqueue.StatusProcessing,
			"attemptlog.0.id": bson.ObjectIdHex(args.ID),
		},
		Change: bson.M{
			"$set": bson.M{"status": workqueue.StatusCompleted},
		},
	}, { //insert the work result
		Collection: "WorkResult",
		DocId:      wkey,
		Insert: WorkResult{
			WorkID:   key,
			Success:  true,
			Revision: args.Revision,
			RevDate:  args.RevDate,
			When:     time.Now(),
		},
	}}

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

		//add the test result to the operation
		ops = append(ops, txn.Operation{
			Collection: "TestResult",
			DocId:      bson.NewObjectId(),
			Insert: TestResult{
				WorkResultID: wkey,
				ImportPath:   out.ImportPath,
				Revision:     args.Revision,
				RevDate:      args.RevDate,
				When:         time.Now(),
				Output:       out.Output,
				Status:       status,
			},
		})
	}

	//run the transaction
	err = ctx.R.Run(ops, nil, nil)
	if err == txn.ErrAborted {
		ctx.Infof("Lost the race inserting result.")
		err = nil
	}

	return
}

//Error is used when there were any errors in building the test
func (Response) Error(req *http.Request, args *rpc.BuilderResponse, resp *rpc.None) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	//create the context
	ctx := httputil.NewContext(req)
	defer ctx.Close()

	//get the key of the work item
	key := bson.ObjectIdHex(args.Key)

	ops := []txn.Operation{{ //make sure we have the given work item
		Collection: "Work",
		DocId:      key,
		Assert: bson.M{
			"status":          workqueue.StatusProcessing,
			"attemptlog.0.id": bson.ObjectIdHex(args.ID),
		},
		Change: bson.M{
			"$set": bson.M{"status": workqueue.StatusCompleted},
		},
	}, { //insert the work result
		Collection: "WorkResult",
		DocId:      bson.NewObjectId(),
		Insert: WorkResult{
			WorkID:   key,
			Success:  false,
			Revision: args.Revision,
			RevDate:  args.RevDate,
			When:     time.Now(),
			Error:    args.Error,
		},
	}}

	//run the transaction
	err = ctx.R.Run(ops, nil, nil)
	if err == txn.ErrAborted {
		ctx.Infof("Lost the race inserting result.")
		err = nil
	}

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
	defer ctx.Close()

	//get the key of the work item
	key := bson.ObjectIdHex(args.Key)

	//create a WorkResult
	w := &WorkResult{
		ID:      bson.NewObjectId(),
		WorkID:  key,
		Success: false,
		When:    time.Now(),
		Error:   args.Error,
	}

	//store it
	if err = ctx.DB.C("WorkResult").Insert(w); err != nil {
		ctx.Errorf("Error storing WorkResult: %v", err)
		return
	}

	//we did it!
	return
}
