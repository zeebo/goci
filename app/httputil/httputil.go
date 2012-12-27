//package httputil provides handler types for requests
package httputil

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"labix.org/v2/mgo/txn"
	"log"
	"net/http"
	"os"
)

//Config is the package level config for httputil.
var Config struct {
	DB     *mgo.Database //Database to use
	Txn    string        //name of transaction collection
	Domain string        //domain name of website

	ContextFunc func(req *http.Request) Context //function to create contexts
}

//internal logger
var logger = log.New(os.Stdout, "", log.Lshortfile)

//set some defaults for the config
func init() {
	Config.Domain = "localhost:9080"
	Config.Txn = "txns"
	Config.ContextFunc = NewContext

	//set the display logs to include filename/num
	log.SetFlags(log.Lshortfile)
}

//Absolute returns an absolute url for a given path.
func Absolute(req string) string {
	return fmt.Sprintf("http://%s%s", Config.Domain, req)
}

//NewContext returns a new context for the given request.
func NewContext(req *http.Request) Context {
	if Config.DB == nil {
		panic("Didn't set httputil.Config.DB before creating contexts")
	}

	db := Config.DB.Session.Clone().DB(Config.DB.Name)

	return Context{
		DB: db,
		R:  txn.NewRunner(db.C(Config.Txn)),
	}
}

//Context represents the set of information a request needs to execute.
type Context struct {
	DB *mgo.Database
	R  *txn.Runner
}

//close closes the context, cleaning up any resources it acquired.
func (c *Context) Close() {
	if c.DB != nil {
		c.DB.Session.Close()
	}
}

//logf pushes the log message into the capped collection with the given format
//and severity.
func (c *Context) logf(severity, format string, items ...interface{}) {
	logger.Output(3, fmt.Sprintf("%s: %s", severity, fmt.Sprintf(format, items...)))
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

//Perform runs the http request through the Handler
func Perform(w http.ResponseWriter, r *http.Request, fn Handler) {
	c := Config.ContextFunc(r)
	defer c.Close()

	//run the handler
	if e := fn(w, r, c); e != nil {
		http.Error(w, e.Message, e.Code)
		c.Errorf("[%d] %s (%v)", e.Code, e.Message, e.Error)
	}
}

//ServeHTTP makes Handler a http.Handler. It allocates and cleans up a context
//and handles errors returned from a handler.
func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Perform(w, r, fn)
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
