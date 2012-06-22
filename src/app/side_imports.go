package main

import (
	_ "env" //env hooks
	"log"
	_ "net/http/pprof" //debug + profiling
)

func init() {
	log.SetFlags(0)
}
