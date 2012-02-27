package main

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
)

func TestMalformedJson(t *testing.T) {
	called := isCalled(false)
	defer setupErrLogger(&called)()

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

const fromGithub = `{
"before": "5aef35982fb2d34e9d9d4502f6ede1072793222d",
  "repository": {
    "url": "http://github.com/defunkt/github",
    "name": "github",
    "description": "You're lookin' at it.",
    "watchers": 5,
    "forks": 2,
    "private": true,
    "owner": {
      "email": "chris@ozmm.org",
      "name": "defunkt"
    }
  },
  "commits": [
    {
      "id": "41a212ee83ca127e3c8cf465891ab7216a705f59",
      "url": "http://github.com/defunkt/github/commit/41a212ee83ca127e3c8cf465891ab7216a705f59",
      "author": {
        "email": "chris@ozmm.org",
        "name": "Chris Wanstrath"
      },
      "message": "okay i give in",
      "timestamp": "2008-02-15T14:57:17-08:00",
      "added": ["filepath.rb"]
    },
    {
      "id": "de8251ff97ee194a289832576287d6f8ad74e3d0",
      "url": "http://github.com/defunkt/github/commit/de8251ff97ee194a289832576287d6f8ad74e3d0",
      "author": {
        "email": "chris@ozmm.org",
        "name": "Chris Wanstrath"
      },
      "message": "update pricing a tad",
      "timestamp": "2008-02-15T14:36:34-08:00"
    }
  ],
  "after": "de8251ff97ee194a289832576287d6f8ad74e3d0",
  "ref": "refs/heads/master"
}`

func TestGoodJson(t *testing.T) {
	defer setupErrLogger(&errorLogger{t})()

	enc := url.Values{
		"payload": {fromGithub},
	}
	body := bytes.NewBufferString(enc.Encode())
	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := NewLoggingRW(t)

	//this shouldn't error
	handlePush(w, req)
}
