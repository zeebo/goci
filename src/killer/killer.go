package main

import (
	"fmt"
	"heroku"
	"log"
	"os"
	"time"
)

const too_long = 90

func clean_processes(app, api string) {
	cl := heroku.New(app, api)
	ps, err := cl.List()
	if err != nil {
		log.Printf("error listing process: %s", err)
		return
	}

	for _, p := range ps {
		if p.Elapsed <= too_long {
			continue
		}

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
		clean_processes(os.Args[0], os.Args[1])
		<-time.After(time.Minute)
	}
}
