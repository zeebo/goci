//package httputil provides handler types for appengine requests
package httputil

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
)

var Config struct {
	DB     *mgo.Database
	Domain string
}

func Absolute(req string) string {
	return fmt.Sprintf("http://%s%s", Config.Domain, req)
}

func NewContext(req *http.Request) Context {
	if Config.DB == nil {
		panic("Set httputil.Config.DB before creating contexts")
	}

	return Context{
		DB: Config.DB.Session.Clone().DB(Config.DB.Name),
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
