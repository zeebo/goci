package main

import "log"

var (
	save_item = make(chan interface{})
)

func init() {
	go save_runner()
}

func save_runner() {
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
		log.Printf("save %s: %#v", coll, it)
	}
}
