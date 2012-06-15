package main

import (
	"log"
	_ "net/http/pprof" //debug + profiling
	_ "env"            //env hooks
)

func init() {
	log.SetFlags(0)
}
