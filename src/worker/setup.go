package worker

import (
	"builder"
	"heroku"
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

	capsize = 100 * Megabyte
)

//collection names
const (
	worklog   = "worklog"
	workqueue = "workqueue"
)

//helper type
type d map[string]interface{}

//run the setup to run the work
func Setup(c Config) {
	//spawn our service goroutines
	go run_test_scheduler()
	go run_run_scheduler()
	go run_saver()

	//spawn the mongo work goroutines
	go run_mgo_work_queue()
	go run_mgo_queue_dispatcher()

	//set up the state changer
	go state_manager()
	change_state <- StateSetup

	//store the config
	config = c

	//set the goroot env var to the default in the env
	setup.GOROOT, builder.GOROOT = config.GOROOT, config.GOROOT

	//build cached heroku client
	hclient = config.BuildHerokuClient()

	//build cached database value
	db = config.BuildMongoDatabase()

	//make sure that mongo is set up properly
	err := db.Run(bson.D{{"create", worklog}, {"size", capsize}, {"capped", true}}, nil)
	if e, ok := err.(*mgo.QueryError); err != nil && (!ok || e.Message != "collection already exists") {
		log.Fatal("error creating collection: ", err)
	}
	log.Println("collection created")

	//set all processing things to false to clean up any old ones
	err = db.C(workqueue).UpdateAll(d{"processing": true}, d{"$set": d{"processing": false}})
	if err != nil && err != mgo.NotFound {
		log.Fatal("error resetting processing values: ", err)
	}

	log.Println("running setup...")
	setup.PrintVars()

	//ensure we have the go tool and vcs in parallel
	var group sync.WaitGroup
	group.Add(2)

	//check the go tool
	go func() {
		if err := setup.EnsureTool(); err != nil {
			log.Fatal(err)
		}
		log.Println("tooling complete")
		group.Done()
	}()

	//check for hg + bzr
	go func() {
		if err := setup.EnsureVCS(); err != nil {
			log.Fatal(err)
		}
		log.Println("vcs complete")
		group.Done()
	}()

	group.Wait()

	log.Println("setup complete. running queue")
	change_state <- StateIdle

	//run our processing queue
	go run_work_queue()
}
