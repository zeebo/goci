package main

import (
	"github.com/zeebo/goci/app/rpc"
	"log"
)

func process(task rpc.BuilderTask) {
	log.Printf("got task: %+v", task)
}
