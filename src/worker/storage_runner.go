package worker

import (
	"log"
	"sync"
)

type ider interface {
	WholeID() string
}

type project_status struct {
	lock       sync.Mutex
	status_map map[string]WorkStatus
}

var (
	save_item    = make(chan *Work)
	status_cache = new(project_status)
)

const (
	cache_len = 50
)

func GetProjectStatus(project string) (WorkStatus, bool) {
	lock.Lock()
	defer lock.Unlock()
	return status_cache.status_map[project]
}

func SetProjectStatus(project string, status WorkStatus) {
	lock.Lock()
	defer lock.Unlock()
	if len(project_status.status_map) > cache_len { //flush cache
		status_cache.status_map = make(map[string]WorkStatus)
	}
	status_cache.status_map[project] = status
}

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

		SetProjectStatus(w.ImportPath, w.Status)
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
