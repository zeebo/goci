package worker

import "launchpad.net/mgo"

type Context struct {
	db *mgo.Database
}

func NewContext() *Context {
	return &Context{
		db: db.Session.Clone().DB(config.Name),
	}
}

func (c *Context) Close() {
	c.db.Session.Close()
}
