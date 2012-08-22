//package tracker provides tracking/announcing of builders and runners
package tracker

import (
	gorpc "code.google.com/p/gorilla/rpc"
	gojson "code.google.com/p/gorilla/rpc/json"
	"errors"
	"github.com/zeebo/goci/app/httputil"
	"github.com/zeebo/goci/app/pinger"
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/app/rpc/client"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"net/http"
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
	http.Handle("/rpc/tracker", s)
}

//Tracker is an rpc for announcing and managing the presence of services
type Tracker struct {
	pinger.Pinger
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

	ctx := httputil.NewContext(req)
	ctx.Infof("Got announce request from %s: %+v", req.RemoteAddr, args)

	//ping them to make sure we can make valid rpc calls
	cl := client.New(args.URL, http.DefaultClient, client.JsonCodec)
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
	key := bson.NewObjectId()
	switch args.Type {
	case "Builder":
		e = &Builder{
			ID:     key,
			GOOS:   args.GOOS,
			GOARCH: args.GOARCH,
			URL:    args.URL,
			Seed:   seed,
		}
	case "Runner":
		e = &Runner{
			ID:     key,
			GOOS:   args.GOOS,
			GOARCH: args.GOARCH,
			URL:    args.URL,
			Seed:   seed,
		}
	default:
		panic("unreachable")
	}

	//TODO(zeebo): check if we have the URL already and grab that key to update

	//save the service in the database
	if err = ctx.DB.C(args.Type).Insert(e); err != nil {
		return
	}

	//return the hex representation of the key
	rep.Key = key.Hex()
	return
}

//Remove removes a service from the tracker.
func (Tracker) Remove(req *http.Request, args *rpc.RemoveArgs, rep *rpc.None) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	//create our context
	ctx := httputil.NewContext(req)
	ctx.Infof("Got a remove request from %s: %+v", req.RemoteAddr, args)

	//get the key from the argument
	key := bson.ObjectIdHex(args.Key)

	//make sure its an entity
	if !isEntity(args.Kind) {
		err = rpc.Errorf("kind is not Builder or Runner")
		return
	}

	//remove it from the database
	err = ctx.DB.C(args.Kind).Remove(bson.M{"_id": key})
	return
}

//Builder is an entity that represents a builder in the tracker.
type Builder struct {
	ID bson.ObjectId `bson:"_id,omitempty"`

	GOOS, GOARCH string
	URL          string

	//Seed is used to distribute work among builders
	Seed int64
}

//Runner is an entity that represents a runner in the tracker.
type Runner struct {
	ID bson.ObjectId `bson:"_id,omitempty"`

	GOOS, GOARCH string
	URL          string

	//Seed is used to distribute work among builders
	Seed int64
}

//ErrNoneAvailable is the error that Lease returns if there are no available
//services matching the criteri
var ErrNoneAvailable = errors.New("no services available")

func baseQuery(db *mgo.Database, GOOS, GOARCH, Type string, Seed int64) (q *mgo.Query) {
	//check for programmer errors
	if !isEntity(Type) {
		panic("type not an entity: " + Type)
	}

	filters := bson.M{}
	//filter on GOOS and GOARCH if they are set
	if GOOS != "" {
		filters["goos"] = GOOS
	}
	if GOARCH != "" {
		filters["goarch"] = GOARCH
	}
	//if we have a Seed value make sure we get one greater than it
	if Seed > 0 {
		filters["seed"] = bson.M{"$gt": Seed}
	}

	//set up the base query
	q = db.C(Type).Find(filters).Limit(1).Sort("seed")

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
	return GOOS + "," + GOARCH + "," + Type
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
func getService(ctx httputil.Context, GOOS, GOARCH, Type string, s interface{}) (err error) {
	//grab the most recent run key
	seed := lastSeeds.get(GOOS, GOARCH, Type)
again:
	ctx.Infof("Finding a %v/%v/%v [%d]", Type, GOOS, GOARCH, seed)
	//run the query
	query := baseQuery(ctx.DB, GOOS, GOARCH, Type, seed)
	err = query.One(s)

	//if we didn't find a match
	if err == mgo.ErrNotFound {

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
func getRunner(ctx httputil.Context, GOOS, GOARCH string) (r *Runner, err error) {
	r = new(Runner)
	err = getService(ctx, GOOS, GOARCH, "Runner", r)
	return
}

//getBuilder grabs a builder from the set of runners matching the given criteria
//in a fashion that attempts to distribute the workload.
func getBuilder(ctx httputil.Context, GOOS, GOARCH string) (b *Builder, err error) {
	b = new(Builder)
	err = getService(ctx, GOOS, GOARCH, "Builder", b)
	return
}

//LeasePair returns a pair of Builder and Runners that can be used to run tests.
//It doesn't let you specify the type of runner you want.
func LeasePair(ctx httputil.Context) (b *Builder, r *Runner, err error) {
	//grab a runner
	r, err = getRunner(ctx, "", "")
	if err != nil {
		ctx.Infof("couldn't lease runner")
		return
	}

	//update the key we're using
	lastSeeds.set("", "", "Runner", r.Seed)

	//grab a builder than can make a build for this runner
	b, err = getBuilder(ctx, r.GOOS, r.GOARCH)
	if err != nil {
		ctx.Infof("couldn't lease builder")
		return
	}

	//update the key we're using
	lastSeeds.set(r.GOOS, r.GOARCH, "Builder", b.Seed)

	return
}
