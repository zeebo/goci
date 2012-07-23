package web

import (
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//dl is a type that represents a path to a file and function to clean it
type dl struct {
	path  string
	clean func()
}

//downloader is a locked map of ids to download paths
type downloader struct {
	paths map[string]dl
	sync.Mutex
}

//newDownloader returns a new downloader ready to go
func newDownloader() *downloader {
	return &downloader{
		paths: map[string]dl{},
	}
}

//Register stores the path for later retieval given the id.
func (d *downloader) Register(download dl) (id string) {
	d.Lock()
	defer d.Unlock()

	//generate an id
	for {
		id = strconv.Itoa(rand.Int())
		if _, ok := d.paths[id]; !ok {
			break
		}
	}

	//store it
	d.paths[id] = download

	//spawn a culler
	go d.Cull(id, 2*time.Minute)

	return
}

//Lookup gets the path for the given id.
func (d *downloader) Lookup(id string) (path string, ok bool) {
	d.Lock()
	defer d.Unlock()
	t, ok := d.paths[id]
	path = t.path
	return
}

//Delete removes the id from the map.
func (d *downloader) Delete(id string) {
	d.Lock()
	defer d.Unlock()
	if t, ok := d.paths[id]; ok {
		delete(d.paths, id)
		t.clean()
	}
}

//Cull waits the TLL and deletes the download given by the id.
func (d *downloader) Cull(id string, ttl time.Duration) {
	<-time.After(ttl)
	d.Delete(id)
}

//downloader serves the file specified by the given id to the client. The file
//is good for only one call.
func (d *downloader) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//grab the id and info
	id := req.URL.Path
	path, ok := d.Lookup(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	//make sure this function only works once
	defer d.Delete(id) //also deletes the file

	//open the file and copy it
	f, err := World.Open(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.Copy(w, f)
	f.Close()
}
