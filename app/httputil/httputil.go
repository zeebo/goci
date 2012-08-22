//package httputil provides handler types for appengine requests
package httputil

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
)

//Config is the package level config for httputil.
var Config struct {
	DB     *mgo.Database
	Domain string
}

//set some defaults for the config
func init() {
	Config.Domain = "localhost:9080"
}

//Absolute returns an absolute url for a given path.
func Absolute(req string) string {
	return fmt.Sprintf("http://%s%s", Config.Domain, req)
}

//NewContext returns a new context for the given request.
func NewContext(req *http.Request) Context {
	if Config.DB == nil {
		panic("Set httputil.Config.DB before creating contexts")
	}

	return Context{
		DB: Config.DB.Session.Clone().DB(Config.DB.Name),
	}
}

//Context represents the set of information a request needs to execute.
type Context struct {
	DB *mgo.Database
}

//close closes the context, cleaning up any resources it acquired.
func (c *Context) Close() {
	c.DB.Session.Close()
}

//logf pushes the log message into the capped collection with the given format
//and severity.
func (c *Context) logf(severity, format string, items ...interface{}) {
	log.Printf("%s: %s", severity, fmt.Sprintf(format, items...))
	c.DB.C("logs").Insert(bson.M{
		"severity": severity,
		"text":     fmt.Sprintf(format, items...),
	})
}

//Errorf pushes an error log message into the system.
func (c *Context) Errorf(format string, items ...interface{}) {
	c.logf("error", format, items...)
}

//Infof pushes an info log message into the system.
func (c *Context) Infof(format string, items ...interface{}) {
	c.logf("info", format, items...)
}

//Handler is a http.Handler for a specific funciton signature.
type Handler func(http.ResponseWriter, *http.Request, Context) *Error

//ServeHTTP makes Handler a http.Handler. It allocates and cleans up a context
//and handles errors returned from a handler.
func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(r)
	defer c.Close()

	//run the handler
	if e := fn(w, r, c); e != nil {
		http.Error(w, e.Message, e.Code)
		c.Errorf("[%d] %s (%v)", e.Code, e.Message, e.Error)
	}
}

//Error is an error type specific to http requests.
type Error struct {
	Error   error
	Message string
	Code    int
}

//Errorf is a helper method to create an Error from the given error.
func Errorf(err error, format string, v ...interface{}) *Error {
	return &Error{err, fmt.Sprintf(format, v...), 500}
}
