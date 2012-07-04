//package tracker provides tracking/announcing of builders and runners
package tracker

import (
	"appengine"
	"appengine/datastore"
	"code.google.com/p/gorilla/rpc"
	"errors"
	"fmt"
	"httputil"
	"net/http"
	"time"
	rpcjson "code.google.com/p/gorilla/rpc/json"
)

func init() {
	//create the rpc server
	s := rpc.NewServer()
	s.RegisterCodec(rpcjson.NewCodec(), "application/json")

	//add the announcer
	s.RegisterService(Announce{}, "")

	//add the announce service to the paths
	http.Handle("/tracker/announce", s)
	http.Handle("/tracker/clean", httputil.Handler(clean))
}

const (
	ttl   = 1 * time.Minute
	retry = 10 * time.Second
)

//clean is a cron job that looks for old or stale services in the datastore and culls them
func clean(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	if err := datastore.RunInTransaction(ctx, clean_transaction, nil); err != nil {
		e = httputil.Errorf(err, "couldn't process clean job")
	}
	return
}

//clean_transaction is the function that clean calls and runs inside a datastore transaction
func clean_transaction(ctx appengine.Context) (err error) {
	//grab all the keys of the things that need to be cleaned
	q := datastore.NewQuery("Service").
		Filter("LastAnnounce < ", time.Now().Sub(ttl)).
		KeysOnly()

	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		return
	}

	err = datastore.DeleteMulti(ctx, keys)
	return
}

//Announce is an rpc for announcing the presence of services to the tracker
type Announce struct{}

//AnnounceArgs is the argument type of the Announce function
type AnnounceArgs struct {
	GOOS, GOARCH string //the goos/goarch of the service
	Type         string //either "Builder" or "Runner"
	URL          string //the url of the service to make rpc calls
}

func (args *AnnounceArgs) verify() (err error) {
	switch {
	case args.GOARCH == "":
		err = errors.New("GOARCH unspecified")
	case args.GOOS == "":
		err = errors.New("GOOS unspecified")
	case args.Type != "Builder" && args.Type != "Runner":
		err = fmt.Errorf("unknown Type: %s", args.Type)
	case args.URL == "":
		err = errors.New("URL unspecified")
	}
	return
}

//AnnounceReply is the reply type of the Announce function
type AnnounceReply struct {
	//the minimum amount of time the service should wait until retrying
	//the announce
	RetryIn time.Duration

	//the amount of time the service will stay active. services are encouraged
	//to announce earlier than the TTL to stay active. 1/2 the TTL should be the
	//minimum amount of time.
	TTL time.Duration
}

//Announce adds the given service into the tracker pool
func (Announce) Announce(req *http.Request, args *AnnounceArgs, rep *AnnounceReply) (err error) {
	if err = args.verify(); err != nil {
		rep.RetryIn = retry
		return
	}
	ctx := appengine.NewContext(req)

	//TODO(zeebo): make a connection to the provided URL
	//and call the Type.Ping method to make sure it exists

	//create the service
	s := &Service{
		GOOS:         args.GOOS,
		GOARCH:       args.GOARCH,
		Type:         args.Type,
		URL:          args.URL,
		LastAnnounce: time.Now(),
		Outstanding:  false,
	}

	//save the new service in the datastore
	if _, err = datastore.Put(ctx, nil, s); err != nil {
		rep.RetryIn = retry
		return
	}

	//send the time to live for the service
	rep.TTL = ttl
	return
}

//Service represents a service in the tracker.
type Service struct {
	GOOS, GOARCH string
	Type         string
	URL          string
	Outstanding  bool
	LastAnnounce time.Time
}

//ErrNoneAvailable is the error that Query returns if there are no available
//services matching the criteri
var ErrNoneAvailable = errors.New("no services available")

//Query returns a URL to a service of the given GOOS, GOARCH and type with no
//outstanding requests and sets that there is an outstanding request for that
//service. If GOOS or GOARCH are the empty string, they are not considered in
//the query.
func Query(ctx appengine.Context, GOOS, GOARCH, Type string) (URL string, err error) {
	tx := func(c appengine.Context) (err error) {
		//set up the base query
		q := datastore.NewQuery("Service").
			Filter("Type = ", Type).
			Filter("LastAnnounce > ", time.Now().Sub(ttl)).
			Filter("Outstanding = ", false).
			Limit(1)

		//filter on GOOS and GOARCH if they are set
		if GOOS != "" {
			q = q.Filter("GOOS = ", GOOS)
		}
		if GOARCH != "" {
			q = q.Filter("GOARCH = ", GOARCH)
		}

		//set up some variables
		var (
			s   = new(Service)
			key *datastore.Key
		)

		//grab the key and value out of the query
		key, err = q.Run(c).Next(s)
		if err == datastore.Done {
			err = ErrNoneAvailable
			return
		}
		if err != nil {
			return
		}

		//attempt to set outstanding to true
		s.Outstanding = true
		_, err = datastore.Put(ctx, key, s)
		if err != nil {
			return
		}

		URL = s.URL
		return
	}
	err = datastore.RunInTransaction(ctx, tx, nil)
	return
}

//QueryAny returns a URL to a service of the given type with the no outstanding
//requests and sets the that there is an outstanding request for that service.
func QueryAny(ctx appengine.Context, Type string) (URL string, err error) {
	URL, err = Query(ctx, "", "", Type)
	return
}
