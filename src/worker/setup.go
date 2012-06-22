package worker

import (
	"builder"
	"heroku"
	"io/ioutil"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"log"
	"setup"
	"sync"
)

var (
	hclient *heroku.Client
	db      *mgo.Database
	config  Config
)

const (
	Byte = 1 << (iota * 10)
	Kilobyte
	Megabyte

	capsize = 75 * Megabyte
)

//collection names
const (
	worklog   = "worklog"
	workqueue = "workqueue"
	testlog   = "testlog"
)

//helper type
type d map[string]interface{}

func create_capped_collection(name string) error {
	err := db.Run(bson.D{{"create", name}, {"size", capsize}, {"capped", true}}, nil)
	if e, ok := err.(*mgo.QueryError); err != nil && (!ok || e.Message != "collection already exists") {
		return err
	}
	return nil
}

var goroutine_spawn sync.Once

//run the setup to run the work
func Setup(c Config) error {
	//spawn our service goroutines once
	goroutine_spawn.Do(func() {
		go run_test_scheduler()
		go run_run_scheduler()
		go run_saver()
		go run_mgo_work_queue()
		go run_mgo_queue_dispatcher()
		go state_manager()
	})

	//set our state as setup
	change_state <- StateSetup

	//store the config
	config = c

	//set the goroot env var to the default in the env
	setup.GOROOT, builder.GOROOT = config.GOROOT, config.GOROOT

	//create a temporary GOPATH for the go-gettable stuff
	dir, err := ioutil.TempDir("", "gopath")
	if err != nil {
		return err
	}
	builder.GOPATH = dir

	//build cached heroku client
	hclient = config.BuildHerokuClient()

	//build cached database value
	db = config.BuildMongoDatabase()

	//make sure that mongo is set up properly
	collections := []string{
		worklog,
		// testlog,
	}
	for _, c := range collections {
		if err := create_capped_collection(c); err != nil {
			return err
		}
		log.Println("created collection:", c)
	}

	//set all processing things to false to clean up any old ones (doesn't scale)
	_, err = db.C(workqueue).UpdateAll(d{"processing": true}, d{"$set": d{"processing": false}})
	if err != nil && err != mgo.ErrNotFound {
		return err
	}

	log.Println("running setup...")
	setup.PrintVars()

	//ensure we have the go tool and vcs in parallel
	var group sync.WaitGroup
	group.Add(2)
	errors := make(chan error, 2)

	//check the go tool
	go func() {
		if err := setup.EnsureTool(); err != nil {
			errors <- err
		}
		group.Done()
	}()

	//check for hg + bzr
	go func() {
		if err := setup.EnsureVCS(); err != nil {
			errors <- err
		}
		group.Done()
	}()

	group.Wait()

	//pull an error out if it occured
	select {
	case err := <-errors:
		return err
	default:
	}

	//no error so we're good!
	change_state <- StateIdle

	go run_work_queue()
	return nil
}
