package worker

import (
	"builder"
	"bytes"
	"encoding/gob"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"time"
)

const queue_size = 100

var (
	work_queue     = make(chan builder.Work)
	mgo_work_queue = make(chan mongoWorkValue)
)

func Schedule(w builder.Work) {
	work_queue <- w
}

type mongoWorkValue struct {
	ID   bson.ObjectId
	Work builder.Work
}

type mongoWork struct {
	ID   bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Work []byte        `json:"-"`

	//some info for fancy display
	ImportPath   string
	NumRevisions int
	RepoPath     string

	//for mongo filtering
	Processing bool
}

//puts items from the work_queue into a mongo stored backend up to a size
func run_mgo_work_queue() {
	var buf bytes.Buffer
	for item := range work_queue {
		buf.Reset()
		enc := gob.NewEncoder(&buf)

		if err := enc.Encode(&item); err != nil {
			log.Println("unable to encode work item:", err)
			continue
		}

		//create the type thats going to go into the database
		mw := mongoWork{
			Work:         buf.Bytes(),
			ImportPath:   item.ImportPath(),
			NumRevisions: len(item.Revisions()),
			RepoPath:     item.RepoPath(),
		}

		//check if the number is bigger than the queue size
		n, err := db.C(workqueue).Find(d{"processing": false}).Count()
		if err != nil {
			log.Println("unable to check current queue size:", err)
			continue
		}
		if n >= queue_size {
			log.Println("too many items in queue to add work item")
			continue
		}

		//add it
		if err := db.C(workqueue).Insert(mw); err != nil {
			log.Println("error adding work item into queue:", err)
		}
	}
}

//takes items from the mongo stored backend and sends them down the mgo_work_queue
func run_mgo_queue_dispatcher() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	//every 10 seconds
	var mw mongoWork
	for _ = range ticker.C {
		//grab some work from the queue
		err := db.C(workqueue).Find(d{"processing": false}).One(&mw)
		if err == mgo.ErrNotFound {
			continue
		}
		if err != nil {
			log.Println("error trying to get a work item:", err)
			continue
		}

		var work builder.Work
		dec := gob.NewDecoder(bytes.NewReader(mw.Work))
		if err := dec.Decode(&work); err != nil {
			log.Printf("error decoding work item %s into interface: %s", mw.ID, err)
			remove_work_item(mw.ID)
			continue
		}

		//set it to processing
		if err := db.C(workqueue).Update(d{"_id": mw.ID}, d{"$set": d{"processing": true}}); err != nil {
			log.Println("error updating the work item to being processed:", err)
			continue
		}

		mgo_work_queue <- mongoWorkValue{Work: work, ID: mw.ID}
	}
}

func remove_work_item(id bson.ObjectId) {
	if err := db.C(workqueue).Remove(d{"_id": id}); err != nil {
		log.Printf("error removing work id %s from database: %s", id, err)
	}
}

func run_work_queue() {
	for work := range mgo_work_queue {
		change_state <- StateRunning
		run_work_item(work)
		change_state <- StateIdle
	}
}

func run_work_item(work mongoWorkValue) {
	log.Println("got work item:", work.Work.RepoPath())
	w, done := new_work(work.Work), make(chan bool)

	defer finish_work(w, work.ID)

	//create the builds for the work
	builds, err := builder.CreateBuilds(work.Work)
	if err != nil {
		w.Error = err.Error()
		return
	}

	go w.wait_for(len(builds), done)
	//build the work struct out to include all the tests
	for _, build := range builds {
		run_build_item(build, w)
	}

	//wait for the work item to be done
	<-done
}

func run_build_item(build builder.Build, w *Work) {
	b, paths := new_build(build, w), build.Paths()
	go b.cleanup(len(paths))

	if err := build.Error(); err != nil {
		b.Error = err.Error()
		b.Passed = false
		return
	}

	for _, bundle := range paths {
		t := new_test(bundle.Path, bundle.Tarball, b, w)
		schedule_test <- t
	}
}

func finish_work(w *Work, id bson.ObjectId) {
	log.Println(w.WholeID(), "clean up")
	w.update_status()
	save_item <- w
	remove_work_item(id)
}
