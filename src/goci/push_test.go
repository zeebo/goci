package main

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func init() {
	//wait for our environment init
	_ = envInit.Value()

	//run a consumer of the results
	//because we have no database connection while testing
	go func() {
		for _ = range resultsChan {
		}
	}()
}

func TestMalformedJson(t *testing.T) {
	called := isCalled(false)
	defer setupErrLogger(&called)()
	defer setupLogger(noOp(false))()

	body := bytes.NewBufferString(`payload={foasdf}`)
	req, err := http.NewRequest("GET", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	w := NewLoggingRW(t)

	//this should error
	handlePush(w, req)

	if !bool(called) {
		t.Error("Did not fail with invalid json")
	}
}

func TestCleanedUp(t *testing.T) {
	w := NewLoggingRW(t)
	l := &sentinalLogger{
		value: "Cleaning up the repository",
	}
	defer setupLogger(l)()
	//a result from github with multiple commits

	json := `{"pusher":{"name":"zeebo","email":"leterip@me.com"},"repository":{"name":"heroku-basic-app",
	"size":136,"has_wiki":true,"created_at":"2012/02/21 11:44:30 -0800","watchers":1,"private":false,
	"url":"https://github.com/zeebo/heroku-basic-app","fork":false,"language":"Go","pushed_at":
	"2012/02/28 13:52:55 -0800","open_issues":0,"has_downloads":true,"homepage":"http://growing-lightning-4944.herokuapp.com/",
	"has_issues":true,"description":"A basic sample app for Heroku","forks":1,"owner":{"name":"zeebo",
	"email":"leterip@me.com"}},"forced":false,"after":"75a462e207e30bff20975c9cbe2055e27ae2b3ea","head_commit":
	{"added":[],"modified":["README.md"],"author":{"name":"jeff","username":"zeebo","email":"leterip@me.com"},
	"timestamp":"2012-02-28T13:52:06-08:00","removed":[],"url":
	"https://github.com/zeebo/heroku-basic-app/commit/75a462e207e30bff20975c9cbe2055e27ae2b3ea","id":
	"75a462e207e30bff20975c9cbe2055e27ae2b3ea","distinct":true,"message":"undo that silly commit"},"deleted":
	false,"ref":"refs/heads/master","commits":[{"added":[],"modified":["README.md"],"author":{"name":"jeff",
	"username":"zeebo","email":"leterip@me.com"},"timestamp":"2012-02-28T13:51:40-08:00","removed":[],"url":
	"https://github.com/zeebo/heroku-basic-app/commit/260e4d35e11232dc68f47a16fb9582383b6b331f","id":
	"260e4d35e11232dc68f47a16fb9582383b6b331f","distinct":true,"message":"Silly commit"},{"added":[],"modified":
	["README.md"],"author":{"name":"jeff","username":"zeebo","email":"leterip@me.com"},"timestamp":
	"2012-02-28T13:52:06-08:00","removed":[],"url":"https://github.com/zeebo/heroku-basic-app/commit/75a462e207e30bff20975c9cbe2055e27ae2b3ea",
	"id":"75a462e207e30bff20975c9cbe2055e27ae2b3ea","distinct":true,"message":"undo that silly commit"}],
	"compare":"https://github.com/zeebo/heroku-basic-app/compare/49719c5...75a462e","before":"49719c5adeb19ca6bbc9eabd2e3cf9dee9eea3ce",
	"created":false}`

	data := url.Values{
		"payload": {json},
	}

	req, err := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handlePush(w, req)

	//lets just try waiting 30 seconds with a 1 second poll
	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout. Never printed")
		case <-time.After(time.Second):
			//check the value
			if l.found {
				return
			}
		}
	}
}
