package worker

import "labix.org/v2/mgo"

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
