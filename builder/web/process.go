package web

import (
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"github.com/zeebo/goci/builder"
	"net/http"
)

//process takes a task and builds the result and either responds to the tracker
//with the build failure output, or forwards a request to the Runner given by
//the task to get the info for the build.
func (b *Builder) process(task rpc.BuilderTask) {
	//build the work item
	builds, revDate, err := b.b.Build(&task.Work)

	//filter out all the builds that have errors
	var errbuilds []builder.Build
	for _, build := range builds {
		if build.Error != "" {
			errbuilds = append(errbuilds, build)
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
		for _, build := range errbuilds {
			resp.BuildErrors = append(resp.BuildErrors, rpc.Output{
				ImportPath: build.ImportPath,
				Output:     build.Error,
			})
		}

		//send it off and ignore the error
		cl := client.New(task.Response, http.DefaultClient, client.JsonCodec)
		if err := cl.Call("Response.Error", resp, new(rpc.None)); err != nil {

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

	//send off to the runner and ignore the error
	cl := client.New(task.Runner, http.DefaultClient, client.JsonCodec)
	if err := cl.Call("RunnerQueue.Push", req, new(rpc.None)); err != nil {

	}
	return
}
