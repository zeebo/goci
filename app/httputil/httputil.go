// +build !goci

//package httputil provides handler types for appengine requests
package httputil

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"net/http"
)

type Handler func(http.ResponseWriter, *http.Request, appengine.Context) *Error

func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if e := fn(w, r, c); e != nil {
		http.Error(w, e.Message, e.Code)
		c.Errorf("%s (%v)", e.Message, e.Error)
	}
}

type Error struct {
	Error   error
	Message string
	Code    int
}

func Errorf(err error, format string, v ...interface{}) *Error {
	return &Error{err, fmt.Sprintf(format, v...), 500}
}

func ToString(key *datastore.Key) (s string) {
	b, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	//remove the `"` surrounding the string to avoid double marshal
	return string(b[1 : len(b)-1])
}

func FromString(key string) *datastore.Key {
	k := new(datastore.Key)
	b, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, k); err != nil {
		panic(err)
	}
	return k
}
