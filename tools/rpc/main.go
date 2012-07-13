package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/zeebo/goci/app/rpc/client"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func check(err error, hint string) {
	if err == nil {
		return
	}

	fmt.Fprintln(os.Stderr, hint, "error:", err)
	os.Exit(1)
}

type rt struct{}

func (rt) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	var buf bytes.Buffer
	io.Copy(io.MultiWriter(&buf, os.Stdout), req.Body)
	fmt.Println("")
	req.Body = ioutil.NopCloser(&buf)
	resp, err = http.DefaultTransport.RoundTrip(req)
	buf.Reset()
	io.Copy(io.MultiWriter(&buf, os.Stdout), resp.Body)
	fmt.Println("")
	resp.Body = ioutil.NopCloser(&buf)
	return
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 3 {
		check(fmt.Errorf("usage: rpc <url> <method> <args>"), "args")
	}
	url, method, arg := args[0], args[1], args[2]
	hcl := &http.Client{
		Transport: rt{},
	}
	cl := client.New(url, hcl, client.JsonCodec)

	var x, y interface{}
	var err error

	err = json.Unmarshal([]byte(arg), &x)
	check(err, "unmar")

	err = cl.Call(method, x, &y)
	check(err, "call")

	var b []byte
	b, err = json.MarshalIndent(y, "", "\t")
	check(err, "marsh")

	fmt.Printf("%s\n", b)
}
