package main

import (
	"github.com/zeebo/goci/app/pinger"
)

func init() {
	if err := rpc_server.RegisterService(pinger.Pinger{}, ""); err != nil {
		panic(err)
	}
}
