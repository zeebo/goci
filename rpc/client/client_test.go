package client

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

//tripper is a RoundTripper that returns the given response and error for every
//request.
type tripper struct {
	resp *http.Response
	err  error
}

//RoundTrip implements the http.RoundTripper interface
func (t tripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.resp, t.err
}

//given_response returns a http.Client that always returns the given response
//and error for every request.
func given_response(resp *http.Response, err error) *http.Client {
	return &http.Client{
		Transport: tripper{resp, err},
	}
}

func TestRoundTrip(t *testing.T) {
	//create the response
	res := []byte(`{"result":2,"error":null,"id":0}`)
	h := given_response(&http.Response{
		Body: ioutil.NopCloser(bytes.NewReader(res)),
	}, nil)

	//create a client that will give us the response
	cl := New("http://localhost/", h, JsonCodec)

	//perform the call
	var x int
	if err := cl.Call("Foo.Bar", 0, &x); err != nil {
		t.Fatal(err)
	}

	//check the result
	if x != 2 {
		t.Fatalf("Expected %d. Got %d", 2, x)
	}
}
