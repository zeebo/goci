package worker

import (
	"builder"
	"heroku"
	"io/ioutil"
	"labix.org/v2/mgo"
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

func create_capped_collection(name string) (err error) {
	info := &mgo.CollectionInfo{
		ForceIdIndex: true,
		Capped:       true,
		MaxBytes:     capsize,
	}
	err = db.C(name).Create(info)

	//fast path: no error
	if err == nil {
		return
	}

	//we have an error, but is it an acceptable one? if so, nil it
	if e, ok := err.(*mgo.QueryError); ok && e.Message == "collection already exists" {
		err = nil
	}

	return
}

var (
	goroutine_spawn sync.Once
	mongo_spawn     sync.Once
)

//run the setup to run the work
func Setup(c Config) error {
	//spawn our service goroutines once
	goroutine_spawn.Do(func() {
		go run_test_scheduler()
		go run_run_scheduler()
		go run_saver()
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

	mongo_spawn.Do(func() {
		go run_mgo_work_queue()
		go run_mgo_queue_dispatcher()
	})

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

	if !config.ReadOnly {
		err = run_setup()
		if err != nil {
			return err
		}
	}

	//no error so we're good!
	change_state <- StateIdle

	go run_work_queue()
	return nil
}

func run_setup() (err error) {
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
	case err = <-errors:
	default:
	}

	return
}
