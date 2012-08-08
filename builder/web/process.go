package web

import (
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"log"
	"net/http"
)

//process takes a task and builds the result and either responds to the tracker
//with the build failure output, or forwards a request to the Runner given by
//the task to get the info for the build.
func (b *Builder) process(task rpc.BuilderTask) {
	log.Printf("Incoming build: %+v", task)

	//build the work item
	builds, revDate, err := b.b.Build(&task.Work)

	//check if we have any errors
	if err != nil {

		//build our response
		resp := &rpc.BuilderResponse{
			Key:   task.Key,
			ID:    task.ID,
			Error: err.Error(),

			//these fields may be zero if we didn't get that far
			Revision: task.Work.Revision,
			RevDate:  revDate,
		}

		log.Printf("Pushing error[%s]: %+v", task.Response, resp)

		//send it off and ignore the error
		cl := client.New(task.Response, http.DefaultClient, client.JsonCodec)
		if err := cl.Call("Response.Error", resp, new(rpc.None)); err != nil {
			//ignored
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
	for _, build := range builds {
		//if the build has an error, then add it to the failures and continue
		//no need to schedule a download
		if build.Error != "" {
			req.WontBuilds = append(req.WontBuilds, rpc.Output{
				ImportPath: build.ImportPath,
				Config:     build.Config,
				Output:     build.Error,
				Type:       rpc.OutputWontBuild,
			})
			continue
		}

		//register the tarball and binary paths with the downloader
		binid := b.dler.Register(dl{
			path:  build.BinaryPath,
			clean: func() { build.CleanBinary() },
		})
		souid := b.dler.Register(dl{
			path:  build.SourcePath,
			clean: func() { build.CleanSource() },
		})

		//add the task with the urls
		req.Tests = append(req.Tests, rpc.RunTest{
			BinaryURL:  b.urlWithPath("/download/" + binid),
			SourceURL:  b.urlWithPath("/download/" + souid),
			ImportPath: build.ImportPath,
			Config:     build.Config,
		})
	}

	log.Printf("Pushing request[%s]: %+v", task.Runner, req)

	//send off to the runner and ignore the error
	cl := client.New(task.Runner, http.DefaultClient, client.JsonCodec)
	if err := cl.Call("RunnerQueue.Push", req, new(rpc.None)); err != nil {

	}
	return
}
