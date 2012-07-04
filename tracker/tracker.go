//package tracker provides tracking/announcing of builders and runners
package tracker

import (
	"appengine"
	"appengine/datastore"
	"errors"
	"httputil"
	"net/http"
	"rpc"
	"time"
	gorpc "code.google.com/p/gorilla/rpc"
	gojson "code.google.com/p/gorilla/rpc/json"
)

func init() {
	//create the rpc server
	s := gorpc.NewServer()
	s.RegisterCodec(gojson.NewCodec(), "application/json")

	//add the announcer
	s.RegisterService(Announce{}, "")

	//add the announce service to the paths
	http.Handle("/tracker", s)
	http.Handle("/tracker/clean", httputil.Handler(clean))
}

const (
	ttl   = 1 * time.Minute
	retry = 10 * time.Second
)

//clean is a cron job that looks for old or stale services in the datastore and culls them
func clean(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	//grab all the keys of the things that need to be cleaned
	q := datastore.NewQuery("Service").
		Filter("LastAnnounce < ", time.Now().Add(-1*ttl)). //LastAnnounce 
		KeysOnly()
	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		e = httputil.Errorf(err, "unable to get keys for cleaning")
		return
	}

	//only delete and log if something is expired
	if len(keys) == 0 {
		return
	}

	//delete them all
	if err = datastore.DeleteMulti(ctx, keys); err != nil {
		e = httputil.Errorf(err, "couldn't delete old keys")
		return
	}

	ctx.Infof("Deleted %d expired services", len(keys))
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

//verify makes sure that the arguments are all specified correctly and returns
//an error that can be encoded over an rpc request.
func (args *AnnounceArgs) verify() (err error) {
	switch {
	case args.GOARCH == "":
		err = rpc.Errorf("GOARCH unspecified")
	case args.GOOS == "":
		err = rpc.Errorf("GOOS unspecified")
	case args.Type != "Builder" && args.Type != "Runner":
		err = rpc.Errorf("unknown Type: %s", args.Type)
	case args.URL == "":
		err = rpc.Errorf("URL unspecified")
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
	defer func() {
		//if we don't have an rpc.Error, encode it as one
		if _, ok := err.(rpc.Error); err != nil && !ok {
			err = rpc.Errorf("%s", err)
		}
	}()

	if err = args.verify(); err != nil {
		rep.RetryIn = retry
		return
	}

	ctx := appengine.NewContext(req)
	ctx.Infof("Got announce request from %s", req.RemoteAddr)

	//TODO(zeebo): make a connection to the provided URL
	//and call the Type.Ping method to make sure it exists
	// cl := client.NewClient(args.URL, urlfetch.Client(ctx), client.JsonCodec)
	// err = cl.Call(fmt.Sprintf("%s.Ping", args.Type), nil, nil)
	// if err != nil {
	// 	rep.RetryIn = retry
	// 	return
	// }

	//create the service entity
	s := &Service{
		GOOS:         args.GOOS,
		GOARCH:       args.GOARCH,
		Type:         args.Type,
		URL:          args.URL,
		LastAnnounce: time.Now(),
		Outstanding:  false,
	}

	//TODO(zeebo): check if we have the URL already and grab that key to update
	key := datastore.NewIncompleteKey(ctx, "Service", nil)

	//save the service in the datastore
	if _, err = datastore.Put(ctx, key, s); err != nil {
		rep.RetryIn = retry
		return
	}

	//send the time to live for the service
	rep.TTL = ttl
	return
}

//Service is an entity that represents a service in the tracker.
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
	//set up the base query
	q := datastore.NewQuery("Service").
		Filter("Type = ", Type).
		Filter("LastAnnounce > ", time.Now().Add(-1*ttl)).
		Filter("Outstanding = ", false).
		Limit(1).
		KeysOnly()

	//filter on GOOS and GOARCH if they are set
	if GOOS != "" {
		q = q.Filter("GOOS = ", GOOS)
	}
	if GOARCH != "" {
		q = q.Filter("GOARCH = ", GOARCH)
	}

	//set up some variables for control flow
	var (
		key   *datastore.Key
		again = errors.New("again")
	)

try_again:
	//grab the key and value out of the query
	key, err = q.Run(ctx).Next(nil)
	if err == datastore.Done {
		err = ErrNoneAvailable
		return
	}
	if err != nil {
		return
	}

	//at this point we grabbed a candidate service
	//now we try in a transaction to recheck that outstanding is false, and if
	//so set it to true.
	tx := func(c appengine.Context) (err error) {
		//make sure outstanding is still false
		s := new(Service)
		err = datastore.Get(c, key, s)
		if err != nil {
			return
		}

		//if it is now outstanding, try to get a new service
		if s.Outstanding {
			err = again
			return
		}

		//attempt to set outstanding to true
		s.Outstanding = true
		_, err = datastore.Put(c, key, s)
		if err != nil {
			return
		}

		URL = s.URL
		return
	}

	//run the transaction
	err = datastore.RunInTransaction(ctx, tx, nil)

	//if it tells us to try because someone else grabbed the key, try again
	if err == again {
		goto try_again
	}

	return
}

//QueryAny returns a URL to a service of the given type with the no outstanding
//requests and sets the that there is an outstanding request for that service.
func QueryAny(ctx appengine.Context, Type string) (URL string, err error) {
	URL, err = Query(ctx, "", "", Type)
	return
}
