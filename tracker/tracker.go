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

	//add the tracker
	s.RegisterService(DefaultTracker, "")

	//add the tracker service to the paths
	http.Handle("/tracker", s)
	http.Handle("/tracker/clean", httputil.Handler(clean))
}

const (
	ttl   = 6 * time.Minute
	retry = 10 * time.Second
)

//clean is a cron job that looks for old or stale services in the datastore and culls them
func clean(w http.ResponseWriter, req *http.Request, ctx appengine.Context) (e *httputil.Error) {
	//TODO(zeebo): make this process smarter. gotta think about leasing builders
	//out and assuming they die at any stage and how to recover.

	//grab all the keys of the things that need to be cleaned
	q := datastore.NewQuery("Service").
		Filter("LastAnnounce < ", time.Now().Add(-1*ttl)).
		Filter("Outstanding = ", false). //don't clean things marked as having something
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

//Tracker is an rpc for announcing and managing the presence of services
type Tracker struct{}

//Set up a DefaultTracker so it can be called without an rpc layer
var DefaultTracker = Tracker{}

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
	//to announce earlier than the TTL to stay active. 1/3 the TTL should be the
	//minimum amount of time.
	TTL time.Duration

	//Key is the datastore key that corresponds to the service if successful
	Key *datastore.Key
}

//Announce adds the given service into the tracker pool
func (Tracker) Announce(req *http.Request, args *AnnounceArgs, rep *AnnounceReply) (err error) {
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
	if key, err = datastore.Put(ctx, key, s); err != nil {
		rep.RetryIn = retry
		return
	}

	//send the time to live for the service
	rep.TTL = ttl
	rep.Key = key
	ctx.Infof("stored at: %s", s.LastAnnounce)
	return
}

//KeepAliveArgs is the argument type of the KeepAlive function
type KeepAliveArgs struct {
	Key *datastore.Key
}

//KeepAliveReply is the reply type of the KeepAlive function
type KeepAliveReply struct {
	//the mininum amount of time the service should wait until retrying the
	//keep alive
	RetryIn time.Duration

	//the amount of time the service will stay active. services are encouraged
	//to send a keep alive earlier than the TTL to stay active. 1/3 the TTL
	//should be the minimum amount of time.
	TTL time.Duration
}

//KeepAlive updates the service so that it stays alive
func (Tracker) KeepAlive(req *http.Request, args *KeepAliveArgs, rep *KeepAliveReply) (err error) {
	defer func() {
		//if we don't have an rpc.Error, encode it as one
		if _, ok := err.(rpc.Error); err != nil && !ok {
			err = rpc.Errorf("%s", err)
		}
	}()

	//check the kind of the key
	if args.Key.Kind() != "Service" {
		err = rpc.Errorf("key does not correspond to a Service")
		return
	}

	//create a context
	ctx := appengine.NewContext(req)
	ctx.Infof("Got keep alive request from %s", req.RemoteAddr)

	//load up the service
	s := new(Service)
	if err = datastore.Get(ctx, args.Key, s); err != nil {
		rep.RetryIn = retry
		return
	}

	//update the LastAnnounce field and save it
	s.LastAnnounce = time.Now()
	if _, err = datastore.Put(ctx, args.Key, s); err != nil {
		rep.RetryIn = retry
		return
	}

	//we're all updated
	rep.TTL = ttl
	ctx.Infof("updated at: %s", s.LastAnnounce)
	return
}

//RemoveArgs is the argument type of the Remove function
type RemoveArgs struct {
	Key *datastore.Key
}

//RemoveReply is the reply type of the Remove function
type RemoveReply struct {
	//the mininum amount of time the service should wait until retrying the
	//remove
	RetryIn time.Duration
}

//Remove removes a service from the tracker.
func (Tracker) Remove(req *http.Request, args *RemoveArgs, rep *RemoveReply) (err error) {
	defer func() {
		//if we don't have an rpc.Error, encode it as one
		if _, ok := err.(rpc.Error); err != nil && !ok {
			err = rpc.Errorf("%s", err)
		}
	}()

	ctx := appengine.NewContext(req)
	ctx.Infof("Got a remove request from %s", req.RemoteAddr)

	//ensure what we have is a service
	if args.Key.Kind() != "Service" {
		err = rpc.Errorf("key is not a service")
		rep.RetryIn = retry
		return
	}

	//delete it
	err = datastore.Delete(ctx, args.Key)
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

//Stale returns if the services LastAnnounce has happened far enough in the past
//to mark it as eligable to be cleaned.
func (s *Service) Stale() bool {
	return time.Now().Sub(s.LastAnnounce) >= ttl
}

//ErrNoneAvailable is the error that Lease returns if there are no available
//services matching the criteri
var ErrNoneAvailable = errors.New("no services available")

//Lease returns a key to a service of the given GOOS, GOARCH and Type with no
//outstanding requests and sets that there is an outstanding request for that
//service. If GOOS or GOARCH are the empty string, they are not considered in
//the query.
func Lease(ctx appengine.Context, GOOS, GOARCH, Type string) (key *datastore.Key, err error) {
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

try_again:
	//set up a sentinel error value for looping
	var again bool

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
		if err = datastore.Get(c, key, s); err != nil {
			return
		}

		//if it is now outstanding, try to get a new service
		if s.Outstanding {
			again = true
			return
		}

		//attempt to set outstanding to true
		s.Outstanding = true
		if _, err = datastore.Put(c, key, s); err != nil {
			return
		}
		return
	}

	//run the transaction
	err = datastore.RunInTransaction(ctx, tx, nil)
	if err != nil {
		key = nil
		return
	}

	//if it tells us to try because someone else grabbed the key, try again
	if again {
		goto try_again
	}

	return
}

//LeaseAny returns a key to a service of the given type with the no outstanding
//requests and sets the that there is an outstanding request for that service.
func LeaseAny(ctx appengine.Context, Type string) (key *datastore.Key, err error) {
	key, err = Lease(ctx, "", "", Type)
	return
}

//Unlease signals to the tracker that the service behind the key is done being
//used and ready to be leased again by another process.
func Unlease(ctx appengine.Context, key *datastore.Key) (err error) {
	//make sure we're unleasing a service
	if key.Kind() != "Service" {
		err = errors.New("key is not a Service")
		return
	}

	//create our transaction
	tx := func(c appengine.Context) (err error) {
		//grab the current state of the 
		s := new(Service)
		if err = datastore.Get(c, key, s); err != nil {
			return
		}

		//make sure it has an Outstanding request.
		//if it doesn't just return
		if !s.Outstanding {
			return
		}

		//set the Outstanding flag to false and store it
		s.Outstanding = false
		if _, err = datastore.Put(c, key, s); err != nil {
			return
		}

		return
	}

	//run the update
	err = datastore.RunInTransaction(ctx, tx, nil)
	return
}
