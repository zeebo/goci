package worker

import "labix.org/v2/mgo"

type Context struct {
	db *mgo.Database
}

func NewContext() *Context {
	var d *mgo.Database
	if db != nil {
		d = db.Session.Clone().DB(config.Name)
	}
	return &Context{
		db: d,
	}
}

func (c *Context) Close() {
	if c.db != nil {
		c.db.Session.Close()
	}
}
