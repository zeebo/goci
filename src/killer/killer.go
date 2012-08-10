package main

import (
	"fmt"
	"heroku"
	"log"
	"os"
	"strings"
	"time"
)

const too_long = 90 // seconds

func clean_processes(app, api string) {
	cl := heroku.New(app, api)
	ps, err := cl.List()
	if err != nil {
		log.Printf("error listing process: %s", err)
		return
	}

	for _, p := range ps {
		//only kill processes running longer than too long
		if p.Elapsed <= too_long {
			continue
		}
		//only kill processes that are in the run command type
		if !strings.HasPrefix(p.Process, "run") {
			continue
		}

		log.Println("killing", p)

		//KEEL IT
		err := cl.Kill(p.Process)
		if err != nil {
			log.Printf("error killing %s: %s", p.Process, err)
		}
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: killer <app> <api key>")
		os.Exit(1)
	}

	for {
		clean_processes(os.Args[1], os.Args[2])
		<-time.After(time.Minute)
	}
}
