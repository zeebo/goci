package main

import (
	"bytes"
	"fmt"
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"github.com/zeebo/goci/environ"
	hrpc "github.com/zeebo/goci/heroku/rpc"
	"github.com/zeebo/goci/tarball"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type LocalWorld interface {
	Create(string, os.FileMode) (io.WriteCloser, error)
	TempDir(string) (string, error)
	Make(environ.Command) environ.Proc
	RemoveAll(string) error
}

var World LocalWorld = environ.New()

//responder is a type that knows about the test environment and can post
//responses to the test manager.
type responder struct {
	url   string
	id    string
	index int
	test  rpc.RunTest
}

//createResponse creates a TestResponse with the given error and output strings
func (r *responder) createResponse(output, err string) *hrpc.TestResponse {
	return &hrpc.TestResponse{
		ID: r.id,
		Output: rpc.Output{
			ImportPath: r.test.ImportPath,
			Config:     r.test.Config,
			Output:     output,
			Error:      err,
		},
	}
}

//post sends the TestResponse to the TestManager
func (r *responder) post(args *hrpc.TestResponse) {
	cl := client.New(r.url, http.DefaultClient, client.JsonCodec)
	cl.Call("RunManager.Post", args, new(rpc.None))
}

//bail is a helper function to post an error
func (r *responder) bail(e interface{}) {
	r.post(r.createResponse("", fmt.Sprint(e)))
}

//success is a helper function to post an output
func (r *responder) success(e interface{}) {
	r.post(r.createResponse(fmt.Sprint(e), ""))
}

//loadTest loads the test field of the responder, returning any errors.
func (r *responder) loadTest() (err error) {
	cl := client.New(r.url, http.DefaultClient, client.JsonCodec)
	args := &hrpc.TestRequest{
		ID:    r.id,
		Index: r.index,
	}
	err = cl.Call("RunManager.Request", args, &r.test)
	return
}

//timeout runs the given proc with a timeout, and returns if the process
//finished in the duration specified.
func timeout(r *responder, p environ.Proc, dur time.Duration) (ok bool) {
	done := make(chan bool, 1)
	if err := p.Start(); err != nil {
		r.bail("error starting command")
		return
	}
	defer p.Kill()

	//start a race
	go func() {
		p.Wait()
		done <- true
	}()

	go func() {
		<-time.After(dur)
		done <- false
	}()

	//see who won
	ok = <-done
	return
}

//turn off all the log flags
func init() {
	log.SetFlags(0)
}

func main() {
	//make sure we have the right number of arguments
	if len(os.Args) != 4 {
		log.Print("usage: runner <url to rpc> <test ID> <test index>")
		return
	}

	//set up the responder
	r := new(responder)
	r.url, r.id = os.Args[1], os.Args[2]

	//parse the index
	s64, err := strconv.ParseInt(os.Args[3], 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	r.index = int(s64)

	//load the test description
	if err := r.loadTest(); err != nil {
		log.Fatal(err)
	}
	//now we can post results back.

	//create a temporary directory for the sources
	sdir, err := World.TempDir("src")
	if err != nil {
		r.bail(err)
		return
	}
	defer World.RemoveAll(sdir)

	//download the sources
	sr, err := http.Get(r.test.SourceURL)
	if err != nil || sr.StatusCode != http.StatusOK {
		r.bail(fmt.Sprintf("%d: %v", sr.StatusCode, err))
		return
	}

	//extract them into the directory
	if err := tarball.Extract(sr.Body, sdir); err != nil {
		r.bail(err)
		return
	}

	//close the body
	sr.Body.Close()

	//create the directory for the binary
	bdir, err := World.TempDir("bin")
	if err != nil {
		r.bail(err)
		return
	}
	defer World.RemoveAll(bdir)

	//create the binary file
	binFile := filepath.Join(bdir, "binary")
	bw, err := World.Create(binFile, 0777)
	if err != nil {
		r.bail(err)
		return
	}

	//download the binary
	br, err := http.Get(r.test.BinaryURL)
	if err != nil || br.StatusCode != http.StatusOK {
		r.bail(fmt.Sprintf("%d: %v", br.StatusCode, err))
		return
	}

	//copy the binary data in and close the file
	io.Copy(bw, br.Body)
	bw.Close()
	br.Body.Close()

	//create the command
	env := []string{
		//copy in some basic env vars if we have them
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("GOROOT=%s", os.Getenv("GOROOT")),
		fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH")),
	}
	var buf bytes.Buffer
	cmd := environ.Command{
		W:    &buf,
		Dir:  sdir,
		Env:  env,
		Path: binFile,
		Args: []string{binFile, "-test.v"},
	}
	proc := World.Make(cmd)

	//only allow the test to run for one minute
	finished := timeout(r, proc, time.Minute)
	if finished {
		r.success(buf.String())
	} else {
		r.bail("test lasted more than 1 minute")
	}
}
