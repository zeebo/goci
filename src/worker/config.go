package worker

import (
	"crypto/rand"
	"fmt"
	"heroku"
	"labix.org/v2/mgo"
	"log"
	"net/url"
	"path"
)

type Config struct {
	ReadOnly  bool
	Debug     bool
	App, Api  string //heroku app/api
	Name, URL string //database name/url
	GOROOT    string //where goroot is
	Host      string //what our host is for building callback urls
}

func (c Config) BuildURL(id string) string {
	p := url.URL{
		Host:   c.Host,
		Path:   path.Join("/bins", id),
		Scheme: "http",
	}
	return p.String()
}

func (c Config) BuildHerokuClient() *heroku.Client {
	return heroku.New(c.App, c.Api)
}

func (c Config) BuildMongoDatabase() *mgo.Database {
	sess, err := mgo.Dial(c.URL)
	if err != nil {
		panic(err)
	}
	return sess.DB(c.Name)
}

func new_id() string {
	const idSize = 10
	var (
		buf [idSize]byte
		n   int
	)
	for n < idSize {
		m, err := rand.Read(buf[n:])
		if err != nil {
			log.Panicf("error generating a random id [%d bytes of %d]: %v", n, idSize, err)
		}
		n += m
	}
	return fmt.Sprintf("%X", buf)
}
