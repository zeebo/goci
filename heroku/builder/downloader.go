package main

import (
	"math/rand"
	"strconv"
	"sync"
)

//defaultDownloader is the downloader we use for keeping track of our binaries.
var defaultDownloader = downloader{items: map[string]string{}}

//downloader is a locked map of ids to file paths.
type downloader struct {
	sync.Mutex
	items map[string]string
}

//Register stores the path for later retieval given the id.
func (d *downloader) Register(path string) (id string) {
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
	d.items[id] = path
	return
}

//Lookup gets the path for the given id.
func (d *downloader) Lookup(id string) (path string, ok bool) {
	d.Lock()
	defer d.Unlock()

	path, ok = d.items[id]
	return
}

//Delete removes the id from the map.
func (d *downloader) Delete(id string) {
	d.Lock()
	defer d.Unlock()

	delete(d.items, id)
}
