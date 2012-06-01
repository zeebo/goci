package main

import "log"

var (
	save_item = make(chan interface{})
)

func init() {
	go save_runner()
}

func save_runner() {
	type ider interface {
		WholeID() string
		GetInfo() TaskInfo
	}
	var id ider
	for it := range save_item {
		var coll string
		switch it.(type) {
		case *Work:
			coll = "Work"
		case *Build:
			coll = "Build"
		case *Test:
			coll = "Test"
		default:
			log.Printf("don't know how to save an item of type %T", it)
			continue
		}
		_ = coll
		id = it.(ider)
		good := id.GetInfo().Error == ""
		log.Println(id.WholeID(), "save. good:", good)
		if !good {
			log.Println(id.WholeID(), "error:", id.GetInfo().Error)
		}
		if t, ok := it.(*Test); ok && good {
			log.Printf("%s output: %q", id.WholeID(), t.Output)
		}
	}
}
