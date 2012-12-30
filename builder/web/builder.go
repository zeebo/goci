package web

import (
	gorpc "github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/zeebo/goci/app/pinger"
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"github.com/zeebo/goci/builder"
	"net/http"
	"net/url"
	"path"
)

//Builder is a type that builds requests sent to it and hosts them temporarily.
type Builder struct {
	b    builder.Builder
	tcl  *client.Client
	base string
	rpc  *gorpc.Server
	bq   rpc.BuilderQueue
	mux  *http.ServeMux
	dler *downloader

	key string
}

//New returns a new web Builder ready to Announce to the given tracker. It
//announces that it is available at `hosted` which should be the full url of
//where this builder resides on the internet.
func New(b builder.Builder, tracker, hosted string) *Builder {
	//create our new builder
	n := &Builder{
		b:    b,
		base: hosted,
		rpc:  gorpc.NewServer(),
		tcl:  client.New(tracker, http.DefaultClient, client.JsonCodec),
		bq:   rpc.NewBuilderQueue(),
		mux:  http.NewServeMux(),
		dler: newDownloader(),
	}

	//register the build service in the rpc
	if err := n.rpc.RegisterService(n.bq, ""); err != nil {
		panic(err)
	}

	//make sure we respond to pings
	if err := n.rpc.RegisterService(pinger.Pinger{}, ""); err != nil {
		panic(err)
	}

	//register the codec
	n.rpc.RegisterCodec(json.NewCodec(), "application/json")

	//add the handlers to our mux
	n.mux.Handle("/", n.rpc)
	n.mux.Handle("/download/", http.StripPrefix("/download/", n.dler))

	//start processing tasks
	go n.run()

	return n
}

//Announce tells the tracker that we're available to build.
func (b *Builder) Announce() (err error) {
	args := &rpc.AnnounceArgs{
		GOOS:   b.b.GOOS(),
		GOARCH: b.b.GOARCH(),
		Type:   "Builder",
		URL:    b.base,
	}
	reply := new(rpc.AnnounceReply)
	if err = b.tcl.Call("Tracker.Announce", args, reply); err != nil {
		return
	}
	b.key = reply.Key
	return
}

//Remove removes this Builder from the tracker.
func (b *Builder) Remove() (err error) {
	args := &rpc.RemoveArgs{
		Key:  b.key,
		Kind: "Builder",
	}
	err = b.tcl.Call("Tracker.Remove", args, new(rpc.None))
	return
}

//ServeHTTP allows the builder to be hosted like any other http.Handler.
func (b *Builder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	b.mux.ServeHTTP(w, req)
}

//run grabs items from the queue and processes them.
func (b *Builder) run() {
	for {
		task := b.bq.Pop()
		b.process(task)
	}
}

//urlWithPath joins the provided path on the end of the base url.
func (b *Builder) urlWithPath(p string) string {
	u, err := url.Parse(b.base)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(u.Path, p)
	return u.String()
}
