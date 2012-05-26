package main

import (
	"builder"
	"encoding/json"
	"github.com/iron-io/iron_mq_go"
	"log"
	"os"
	"setup"
	"sync"
	"time"
)

type Message struct {
	Results []builder.Report
	Work    string
	Error   string
}

func send_to_queue(in chan Message, queue *ironmq.Queue) {
	for m := range in {
		b, err := json.Marshal(m)
		if err != nil {
			log.Println("error marshalling response:", err)
			continue
		}
		log.Println("pushing:", string(b))
		_, err = queue.Push(string(b))
		if err != nil {
			log.Println("error pushing to queue:", err)
		}
	}
}

func pull_from_queue(queue *ironmq.Queue, wait time.Duration, out chan builder.Work) {
	for {
		<-time.After(wait)

		m, err := queue.Get()

		switch err {
		case ironmq.EmptyQueue:
			continue
		case nil:
		default:
			log.Println("error polling queue:", err)
			continue
		}

		w, err := builder.Unserialize(m.Body)
		if err != nil {
			log.Println("error unserializing work:", err)
			continue
		}

		out <- w
		m.Delete()
	}
}

func do_setup() {
	//ensure tool/vcs at the same time
	var group sync.WaitGroup
	group.Add(2)

	builder.GOROOT = setup.GOROOT

	go func() {
		if err := setup.EnsureTool(); err != nil {
			log.Fatal(err)
		}
		group.Done()
	}()

	go func() {
		if err := setup.EnsureVCS(); err != nil {
			log.Fatal(err)
		}
		group.Done()
	}()

	group.Wait()
}

func work_loop(in chan builder.Work, out chan Message) {
	for {
		w := <-in
		wm, err := builder.Serialize(w)
		if err != nil {
			log.Fatal("invalid work item:", err)
			continue
		}

		rep, err := builder.Run(w)
		m := Message{
			Results: rep,
			Work:    wm,
		}
		if err != nil {
			m.Error = err.Error()
		}

		out <- m
	}
}

func main() {
	do_setup()

	client := ironmq.NewClient(
		os.Getenv("IRON_MQ_PROJECT_ID"),
		os.Getenv("IRON_MQ_TOKEN"),
		ironmq.IronAWSUSEast,
	)
	in, out := client.Queue("work_in"), client.Queue("results")

	work := make(chan builder.Work)
	res := make(chan Message)
	go pull_from_queue(in, 5*time.Second, work)
	go send_to_queue(res, out)

	//just run 1 worker for now
	work_loop(work, res)
}
