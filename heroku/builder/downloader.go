package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//downloadType represents a string and a function to clean when removed.
type downloadType struct {
	path  string
	clean func()
}

//defaultDownloader is the downloader we use for keeping track of our binaries.
var defaultDownloader = &downloader{items: map[string]downloadType{}}

//downloader is a locked map of ids to file paths.
type downloader struct {
	sync.Mutex
	items map[string]downloadType
}

//Register stores the path for later retieval given the id.
func (d *downloader) Register(path string, clean func()) (id string) {
	d.Lock()
	defer d.Unlock()

	//generate an id for this path
	for {
		id = strconv.Itoa(rand.Int())
		if _, ok := d.items[id]; !ok {
			break
		}
	}

	//store it
	d.items[id] = downloadType{path, clean}

	//spawn a culler to remove the download after 2 minutes
	go d.Cull(id, 2*time.Minute)

	return
}

//Lookup gets the path for the given id.
func (d *downloader) Lookup(id string) (path string, ok bool) {
	d.Lock()
	defer d.Unlock()
	t, ok := d.items[id]
	path = t.path
	return
}

//Delete removes the id from the map.
func (d *downloader) Delete(id string) {
	d.Lock()
	defer d.Unlock()
	if t, ok := d.items[id]; ok {
		delete(d.items, id)
		t.clean()
	}
}

//Cull waits the TLL and deletes the download given by the id.
func (d *downloader) Cull(id string, ttl time.Duration) {
	<-time.After(ttl)
	d.Delete(id)
}

//download serves the file specified by the defaultDownloader and the given id
//to the client. It is good for only one call.
func download(w http.ResponseWriter, req *http.Request) {
	//grab the id and info
	id := req.FormValue(":id")
	path, ok := defaultDownloader.Lookup(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	//make sure the function only works once and clean up after ourselves
	defer defaultDownloader.Delete(id) //also deletes the file

	//open the file
	f, err := World.Open(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	//copy it down the pipe
	_, err = io.Copy(w, f)
	if err != nil {
		log.Println("error copying download:", err)
	}
}
