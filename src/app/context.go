package main

import (
	"log"
	"net/http"

	"github.com/gorilla/sessions"

	"worker"
)

type Context struct {
	Session *sessions.Session
	Meta    *Meta
	Data    d
	State   worker.State
	Context *worker.Context
}

func init_context(req *http.Request) (c *Context) {
	session, err := store.Get(req, appname)
	if err != nil {
		log.Println("store err:", err)
	}
	c = &Context{
		Session: session,
		Meta:    base_meta.Dup(),
		Data:    d{},
		State:   worker.GetState(),
		Context: worker.NewContext(),
	}

	if req.URL.Path != "/" {
		c.Meta.Nav.SetActive(req.URL.Path)
	}

	return
}

func (c *Context) Save(req *http.Request, w http.ResponseWriter) {
	c.Session.Save(req, w)
}

func (c *Context) Close() {
	//perform any cleanup of resources acquired in init_context
	c.Context.Close()
}

//Called in the base template to flatten all the data into one dictionary
//to make accessing elements not have to go through the .Data attribute
func (c *Context) Tmpl() (ret d) {
	ret = d{}
	for key, val := range c.Data {
		ret[key] = val
	}
	ret["Meta"], ret["Session"], ret["State"] = c.Meta, c.Session, c.State
	return ret
}

func (c *Context) Set(key string, val interface{}) {
	c.Data[key] = val
}

func (c *Context) Merge(ex d) {
	for key, val := range ex {
		c.Data[key] = val
	}
}

func (c *Context) Get(key string, def interface{}) interface{} {
	if v, ex := c.Data[key]; ex {
		return v
	}
	return def
}
