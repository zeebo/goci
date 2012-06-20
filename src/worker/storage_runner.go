package worker

import "log"

type ider interface {
	WholeID() string
}

var (
	save_item = make(chan *Work)
)

func run_saver() {
	for w := range save_item {
		good := w.Error == ""
		log.Println(w.WholeID(), "save. good:", good)
		if !good {
			log.Printf("%s error: %q", w.WholeID(), cap(w.Error, 50))
		}
		w.Freeze()

		//perform the save
		if err := db.C(worklog).Insert(w); err != nil {
			log.Printf("%s error saving: %s", w.WholeID(), err)
		}
	}
}

func cap(s string, max int) (v string) {
	if max > len(s) {
		max = len(s)
	}
	v = s[:max]
	return
}

//helper load function that auto unthaws things
func load(sel interface{}) (w *Work, err error) {
	err = db.C(worklog).Find(sel).One(&w)
	if err == nil {
		w.Thaw()
	}
	return
}
