package client

import (
	"github.com/gorilla/rpc/json"
	"io"
)

//jsonCodec wraps the package level functions in the rpc/json package to create
//a client codec.
type jsonCodec struct{}

func (jsonCodec) EncodeRequest(method string, args interface{}) ([]byte, error) {
	return json.EncodeClientRequest(method, args)
}

func (jsonCodec) DecodeResponse(r io.Reader, reply interface{}) error {
	return json.DecodeClientResponse(r, reply)
}

func (jsonCodec) ContentType() string {
	return "application/json"
}

//JsonCodec is a client Codec for interacting with services using the rpc/json
//codec.
var JsonCodec Codec = jsonCodec{}
