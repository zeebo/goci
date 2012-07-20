// +build !goci

//package tracker provides tracking/announcing of builders and runners
package tracker

import (
	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	gorpc "code.google.com/p/gorilla/rpc"
	gojson "code.google.com/p/gorilla/rpc/json"
	"errors"
	"httputil"
	"math/rand"
	"net/http"
	"pinger"
	"rpc"
	"rpc/client"
	"strings"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func init() {
	//create the rpc server
	s := gorpc.NewServer()
	s.RegisterCodec(gojson.NewCodec(), "application/json")

	//add the tracker
	s.RegisterService(DefaultTracker, "")

	//add the tracker service to the paths
	http.Handle("/tracker", s)
}

//Tracker is an rpc for announcing and managing the presence of services
type Tracker struct {
	pinger.Pinger //a Tracker repsonds to ping
}

//Set up a DefaultTracker so it can be called without an rpc layer
var DefaultTracker = Tracker{}

//verify makes sure that the arguments are all specified correctly and returns
//an error that can be encoded over an rpc request.
func verify(args *rpc.AnnounceArgs) (err error) {
	switch {
	case args.GOARCH == "":
		err = rpc.Errorf("GOARCH unspecified")
	case args.GOOS == "":
		err = rpc.Errorf("GOOS unspecified")
	case !isEntity(args.Type):
		err = rpc.Errorf("unknown Type: %s", args.Type)
	case args.URL == "":
		err = rpc.Errorf("URL unspecified")
	}
	return
}

func isEntity(kind string) bool {
	switch kind {
	case "Builder", "Runner":
		return true
	}
	return false
}

//Announce adds the given service into the tracker pool
func (Tracker) Announce(req *http.Request, args *rpc.AnnounceArgs, rep *rpc.AnnounceReply) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	if err = verify(args); err != nil {
		return
	}

	ctx := appengine.NewContext(req)
	ctx.Infof("Got announce request from %s", req.RemoteAddr)

	//ping them to make sure we can make valid rpc calls
	cl := client.New(args.URL, urlfetch.Client(ctx), client.JsonCodec)
	err = cl.Call("Pinger.Ping", nil, new(rpc.None))
	if err != nil {
		ctx.Infof("Failed to Ping the announce.")
		return
	}

	//create the entity
	var e interface{}

	//make sure we have a nonzero seed
	var seed int64
	for seed == 0 {
		seed = rand.Int63()
	}
	switch args.Type {
	case "Builder":
		e = &Builder{
			GOOS:   args.GOOS,
			GOARCH: args.GOARCH,
			URL:    args.URL,
			Seed:   seed,
		}
	case "Runner":
		e = &Runner{
			GOOS:   args.GOOS,
			GOARCH: args.GOARCH,
			URL:    args.URL,
			Seed:   seed,
		}
	default:
		panic("unreachable")
	}

	//TODO(zeebo): check if we have the URL already and grab that key to update
	key := datastore.NewIncompleteKey(ctx, args.Type, nil)

	//save the service in the datastore
	if key, err = datastore.Put(ctx, key, e); err != nil {
		return
	}

	//set the key for the service
	rep.Key = httputil.ToString(key)
	return
}

//Remove removes a service from the tracker.
func (Tracker) Remove(req *http.Request, args *rpc.RemoveArgs, rep *rpc.None) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	//create our context
	ctx := appengine.NewContext(req)
	ctx.Infof("Got a remove request from %s", req.RemoteAddr)

	//grab the key
	key := httputil.FromString(args.Key)

	//ensure what we have is a service
	if !isEntity(key.Kind()) {
		err = rpc.Errorf("key is not a builder or runner")
		return
	}

	//delete it
	err = datastore.Delete(ctx, key)
	return
}

//Builder is an entity that represents a builder in the tracker.
type Builder struct {
	GOOS, GOARCH string
	URL          string

	//Seed is used to distribute work among builders
	Seed int64
}

//Runner is an entity that represents a runner in the tracker.
type Runner struct {
	GOOS, GOARCH string
	URL          string

	//Seed is used to distribute work among builders
	Seed int64
}

//ErrNoneAvailable is the error that Lease returns if there are no available
//services matching the criteri
var ErrNoneAvailable = errors.New("no services available")

func baseQuery(GOOS, GOARCH, Type string, Seed int64) (q *datastore.Query) {
	//check for programmer errors
	if !isEntity(Type) {
		panic("type not an entity: " + Type)
	}

	//set up the base query
	q = datastore.NewQuery(Type).Limit(1).Order("Seed")

	//filter on GOOS and GOARCH if they are set
	if GOOS != "" {
		q = q.Filter("GOOS = ", GOOS)
	}
	if GOARCH != "" {
		q = q.Filter("GOARCH = ", GOARCH)
	}

	//if we have a Seed value make sure we get one greater than it
	if Seed > 0 {
		q = q.Filter("Seed > ", Seed)
	}
	return
}

//seeds is a locked map of strings to seed values.
type seeds struct {
	c map[string]int64
	sync.Mutex
}

//lastSeeds is a mapping of entity types to the last seed value seen of that
//type so that we attempt to distribute load across the services.
var lastSeeds = &seeds{c: map[string]int64{}}

//key returns the key used in the map for the set of constrains.
func (s *seeds) key(GOOS, GOARCH, Type string) string {
	return strings.Join([]string{GOOS, GOARCH, Type}, ",")
}

//get looks up the cached seed value for the given set of constraints.
func (s *seeds) get(GOOS, GOARCH, Type string) (r int64) {
	s.Lock()
	defer s.Unlock()

	r = s.c[s.key(GOOS, GOARCH, Type)]
	return
}

//set sets the cached seed value for the given set of constraints.
func (s *seeds) set(GOOS, GOARCH, Type string, v int64) {
	s.Lock()
	defer s.Unlock()

	s.c[s.key(GOOS, GOARCH, Type)] = v
	return
}

//getService is a helper function that abstracts the logic of grabbing a service
//with a key greater than the one given, and looping back to zero if one wasn't
//found.
func getService(ctx appengine.Context, GOOS, GOARCH, Type string, s interface{}) (key *datastore.Key, err error) {
	//grab the most recent run key
	seed := lastSeeds.get(GOOS, GOARCH, Type)
again:
	ctx.Infof("Finding a %v/%v/%v [%d]", Type, GOOS, GOARCH, seed)
	//run the query
	query := baseQuery(GOOS, GOARCH, Type, seed)
	key, err = query.Run(ctx).Next(s)

	//if we didn't find a match
	if err == datastore.Done {

		//try again if we're limiting on the seed
		if seed > 0 {
			seed = 0
			goto again
		}

		//there just arent any
		err = ErrNoneAvailable
	}

	return
}

//getRunner grabs a runner from the set of runners matching the given criteria
//in a fashion that attempts to distribute the workload.
func getRunner(ctx appengine.Context, GOOS, GOARCH string) (key *datastore.Key, r *Runner, err error) {
	r = new(Runner)
	key, err = getService(ctx, GOOS, GOARCH, "Runner", r)
	return
}

//getBuilder grabs a builder from the set of runners matching the given criteria
//in a fashion that attempts to distribute the workload.
func getBuilder(ctx appengine.Context, GOOS, GOARCH string) (key *datastore.Key, b *Builder, err error) {
	b = new(Builder)
	key, err = getService(ctx, GOOS, GOARCH, "Builder", b)
	return
}

//LeasePair returns a pair of Builder and Runners that can be used to run tests.
//It doesn't let you specify the type of runner you want.
func LeasePair(ctx appengine.Context) (builder, runner *datastore.Key, b *Builder, r *Runner, err error) {
	//grab a runner
	runner, r, err = getRunner(ctx, "", "")
	if err != nil {
		ctx.Infof("couldn't lease runner")
		return
	}

	//update the key we're using
	lastSeeds.set("", "", "Runner", r.Seed)

	//grab a builder than can make a build for this runner
	builder, b, err = getBuilder(ctx, r.GOOS, r.GOARCH)
	if err != nil {
		ctx.Infof("couldn't lease builder")
		return
	}

	//update the key we're using
	lastSeeds.set(r.GOOS, r.GOARCH, "Builder", b.Seed)

	return
}
