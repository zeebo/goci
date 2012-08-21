// +build !goci

//package httputil provides handler types for appengine requests
package httputil

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
)

var baseDb *mgo.Database

func Init(db *mgo.Database) {
	baseDb = db
}

func NewContext(req *http.Request) Context {
	if baseDb == nil {
		panic("httputil.Init not called before creating contexts")
	}

	return Context{
		DB: baseDb.Session.Clone().DB(baseDb.Name),
	}
}

type Context struct {
	DB *mgo.Database
}

func (c *Context) close() {
	c.DB.Session.Close()
}

func (c *Context) logf(severity, format string, items ...interface{}) {
	c.DB.C("logs").Insert(bson.M{
		"severity": severity,
		"text":     fmt.Sprintf(format, items...),
	})
}

func (c *Context) Errorf(format string, items ...interface{}) {
	c.logf("error", format, items...)
}

func (c *Context) Infof(format string, items ...interface{}) {
	c.logf("info", format, items...)
}

type Handler func(http.ResponseWriter, *http.Request, Context) *Error

func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(r)
	defer c.close()
	if e := fn(w, r, c); e != nil {
		http.Error(w, e.Message, e.Code)
		c.Errorf("[%d] %s (%v)", e.Code, e.Message, e.Error)
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
