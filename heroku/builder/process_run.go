package main

import (
	"github.com/zeebo/goci/app/rpc"
	"log"
)

func process_run(task rpc.RunnerTask) {
	log.Println("running task:", task)
}
