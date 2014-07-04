package main

import (
	"log"

	_ "env"            //env hooks
	_ "net/http/pprof" //debug + profiling
)

func init() {
	log.SetFlags(0)
}
