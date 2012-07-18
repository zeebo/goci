package main

import (
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"github.com/zeebo/goci/builder"
	"net/http"
	"net/url"
)

//defaultBuilder is the builder we use and created by the setup function.
var defaultBuilder builder.Builder

//baseUrl is the url we use for downloading binaries and tarballs
var baseUrl *url.URL

//parse our baseUrl in from the environment
func init() {
	var err error
	baseUrl, err = url.Parse(env("BASE_URL", "http://builder.goci.me"))
	if err != nil {
		bail(err)
	}
}

//urlWithPath makes a copy of the baseUrl and sets the path to the provided path
//and returns the string representation
func urlWithPath(path string) string {
	urlCopy := *baseUrl
	urlCopy.Path = path
	return urlCopy.String()
}

//process takes a task and builds the result and either responds to the tracker
//with the build failure output, or forwards a request to the Runner given by
//the task to get the info for the build.
func process(task rpc.BuilderTask) {
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

		//send it off and ignore the error
		cl := client.New(task.Response, http.DefaultClient, client.JsonCodec)
		cl.Call("Response.Error", resp, new(rpc.None))
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
		binid := defaultDownloader.Register(b.BinaryPath)
		souid := defaultDownloader.Register(b.SourcePath)

		//add the task with the urls
		req.Tasks = append(req.Tasks, rpc.RunTest{
			BinaryURL:  urlWithPath("/download/" + binid),
			SourceURL:  urlWithPath("/download/" + souid),
			ImportPath: b.ImportPath,
		})
	}

	//send off to the runner and ignore the error
	cl := client.New(task.Runner, http.DefaultClient, client.JsonCodec)
	cl.Call("RunnerQueue.Push", req, new(rpc.None))
	return
}
