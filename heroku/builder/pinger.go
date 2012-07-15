package main

import "github.com/zeebo/goci/app/pinger"

func init() {
	if err := rpcServer.RegisterService(pinger.Pinger{}, ""); err != nil {
		bail(err)
	}
}
