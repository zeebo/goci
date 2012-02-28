package main

import (
	"bytes"
	"net/http"
	"testing"
)

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
