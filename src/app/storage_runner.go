package main

import "log"

type ider interface {
	WholeID() string
	GetInfo() TaskInfo
}

var (
	save_item = make(chan ider)
)

func run_saver() {
	type freezer interface {
		Freeze()
	}

	for id := range save_item {
		var coll string
		switch id.(type) {
		case *Work:
			coll = "Work"
		case *Build:
			coll = "Build"
		case *Test:
			coll = "Test"
		default:
			log.Printf("don't know how to save an item of type %T", id)
			continue
		}

		good := id.GetInfo().Error == ""
		log.Println(id.WholeID(), "save. good:", good)
		if !good {
			log.Printf("%s error: %q", id.WholeID(), id.GetInfo().Error)
		}
		if t, ok := id.(*Test); ok && good {
			log.Printf("%s passed: %v output: %q", id.WholeID(), t.Passed, t.Output)
		}

		//see if it needs to freeze before storage
		if p, ok := id.(freezer); ok {
			p.Freeze()
		}

		//perform the save
		if err := db.C(coll).Insert(id); err != nil {
			log.Printf("%s error saving: %s", id.WholeID(), err)
		}
	}
}

//helper load function that auto unthaws things
func load(coll string, sel interface{}, it interface{}) (err error) {
	type thawer interface {
		Thaw()
	}
	err = db.C(coll).Find(sel).One(it)
	if t, ok := it.(thawer); ok && err == nil {
		t.Thaw()
	}
	return
}
