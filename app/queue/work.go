// +build !goci

package queue

import (
	"appengine/datastore"
	"reflect"
	"rpc"
	"time"
)

//Work is the datastore entity representing the work item that came in
type Work struct {
	Work    rpc.Work  `datastore:"-"` //the parsed and distilled work item
	Data    string    //the raw data that came in
	Created time.Time //when the item was received
}

var wtype = reflect.TypeOf(Work{})

//dispatch sends all the fields that arent part of a Work to channel 1 first
//and then all the fields that are part of a Work to channel 2.
func dispatch(c <-chan datastore.Property, ch1, ch2 chan<- datastore.Property) {
	var buf []datastore.Property
	for v := range c {
		if _, ok := wtype.FieldByName(v.Name); ok {
			buf = append(buf, v)
		} else {
			ch1 <- v
		}
	}
	close(ch1)
	for _, v := range buf {
		ch2 <- v
	}
	close(ch2)
}

//concat takes a list of channels and sends them in order down the first.
func concat(c chan<- datastore.Property, chs ...<-chan datastore.Property) {
	for _, ch := range chs {
		for v := range ch {
			c <- v
		}
	}
	close(c)
}

//consume reads from a channel until it is closed.
func consume(ch <-chan datastore.Property) {
	for _ = range ch {
	}
}

//Load implements the load portion of the PropertyLoadSaver interface.
func (w *Work) Load(c <-chan datastore.Property) (err error) {
	//buffer and order the results so Work comes first
	ch1, ch2 := make(chan datastore.Property), make(chan datastore.Property)
	done := make(chan bool)
	go func() { dispatch(c, ch1, ch2); close(done) }()

	//consume everything so nothing is leaked
	defer consume(ch2)
	defer consume(ch1)

	if err = datastore.LoadStruct(&w.Work, ch1); err != nil {
		return
	}
	if err = datastore.LoadStruct(w, ch2); err != nil {
		return
	}

	//wait for the ordering to finish
	<-done
	return
}

//Save implements the save portion of the PropertyLoadSaver interface.
func (w *Work) Save(c chan<- datastore.Property) (err error) {
	ch1, ch2 := make(chan datastore.Property), make(chan datastore.Property)

	done := make(chan bool)
	go func() { concat(c, ch1, ch2); close(done) }()

	if err = datastore.SaveStruct(&w.Work, ch1); err != nil {
		return
	}
	if err = datastore.SaveStruct(w, ch2); err != nil {
		return
	}

	//wait for concat
	<-done
	return
}

//TaskInfo is an entity that stores when a task was sent out so the cron can
//readd tasks that have expired.
type TaskInfo struct {
	Key     string
	Created time.Time
}
