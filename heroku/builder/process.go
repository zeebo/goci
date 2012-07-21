package main

import (
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"github.com/zeebo/goci/builder"
	"log"
	"net/http"
)

//process takes a task and builds the result and either responds to the tracker
//with the build failure output, or forwards a request to the Runner given by
//the task to get the info for the build.
func process(task rpc.BuilderTask) {
	log.Printf("Builder processing %+v", task)

	//build the work item
	builds, revDate, err := defaultBuilder.Build(&task.Work)

	//filter out all the builds that have errors
	var errbuilds []builder.Build
	for _, b := range builds {
		if b.Error != "" {
			errbuilds = append(errbuilds, b)
		}
	}

	//check if we have any errors
	if err != nil || len(errbuilds) > 0 {

		//build our response
		resp := &rpc.BuilderResponse{
			Key: task.Key,
			ID:  task.ID,
		}
		if err != nil {
			resp.Error = err.Error()
		}
		for _, b := range errbuilds {
			resp.BuildErrors = append(resp.BuildErrors, rpc.Output{
				ImportPath: b.ImportPath,
				Output:     b.Error,
			})
		}

		log.Printf("Builder error response: %+v", resp)

		//send it off and ignore the error
		cl := client.New(task.Response, http.DefaultClient, client.JsonCodec)
		if err := cl.Call("Response.Error", resp, new(rpc.None)); err != nil {
			log.Println("Error sending build failure response:", err)
		}
		return
	}

	//build the runner request
	req := &rpc.RunnerTask{
		Key:      task.Key,
		ID:       task.ID,
		Revision: task.Work.Revision,
		RevDate:  revDate,
		Response: task.Response,
	}
	for _, b := range builds {
		//register the tarball and binary paths with the downloader
		binid := defaultDownloader.Register(b.BinaryPath, func() { b.CleanBinary() })
		souid := defaultDownloader.Register(b.SourcePath, func() { b.CleanSource() })

		//add the task with the urls
		req.Tests = append(req.Tests, rpc.RunTest{
			BinaryURL:  urlWithPath(reverse("download", "id", binid)),
			SourceURL:  urlWithPath(reverse("download", "id", souid)),
			ID:         task.ID,
			ImportPath: b.ImportPath,
			Config:     b.Config,
		})
	}

	log.Printf("Builder forwarding to runner: %+v", req)

	//send off to the runner and ignore the error
	cl := client.New(task.Runner, http.DefaultClient, client.JsonCodec)
	if err := cl.Call("RunnerQueue.Push", req, new(rpc.None)); err != nil {
		log.Println("Error sending to runner queue:", err)
	}
	return
}
