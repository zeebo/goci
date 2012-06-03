package main

import (
	"heroku"
	"os"
)

func need_env(key string) (val string) {
	if val = os.Getenv(key); val == "" {
		panic("key not found: " + key)
	}
	return
}

var hclient = heroku.New(need_env("APPNAME"), need_env("APIKEY"))
