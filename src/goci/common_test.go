package main

import (
	"bytes"
	"net/http"
	"testing"
)

//simple Logger that fails the test on any call.
type errorLogger struct {
	t *testing.T
}

func (e errorLogger) Fatal(v ...interface{})                 { e.t.Error(v...) }
func (e errorLogger) Printf(format string, v ...interface{}) { e.t.Errorf(format, v...) }
func (e errorLogger) Println(v ...interface{})               { e.t.Error(v...) }

//simple logger that just sets if it gets called
type isCalled bool

func (i *isCalled) Fatal(v ...interface{})                 { *i = true }
func (i *isCalled) Printf(format string, v ...interface{}) { *i = true }
func (i *isCalled) Println(v ...interface{})               { *i = true }

//simple type that acts as a http.ResponseWriter. Errors if Header() is called
//after WriteHeader
type bufferedWriter struct {
	t       *testing.T
	body    bytes.Buffer
	headers http.Header
	status  int
	wrote   bool
}

func NewLoggingRW(t *testing.T) *bufferedWriter {
	return &bufferedWriter{
		t:       t,
		headers: http.Header{},
	}
}

func (b *bufferedWriter) Header() http.Header {
	if b.wrote {
		b.t.Errorf("Header called after headers were written")
		return nil
	}
	return b.headers
}

func (b *bufferedWriter) Write(p []byte) (int, error) {
	if !b.wrote {
		b.wrote = true
		b.status = http.StatusOK
	}

	return b.body.Write(p)
}

func (b *bufferedWriter) WriteHeader(status int) {
	b.status = status
	b.wrote = true
}

type lambda func()

func setupErrLogger(l Logger) lambda {
	oldErrLogger := errLogger
	errLogger = l
	return func() {
		errLogger = oldErrLogger
	}
}
