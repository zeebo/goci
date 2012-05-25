package main

import (
	"bufio"
	"github.com/iron-io/iron_mq_go"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func env() (err error) {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()
	b := bufio.NewReader(f)
	var line string

	for {
		line, err = b.ReadString('\n')
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return
		}

		chunks := strings.SplitN(line, "=", 2) //split into 2 things
		os.Setenv(chunks[0], strings.TrimSpace(chunks[1]))
	}
	return
}

func poll(queue *ironmq.Queue, wait time.Duration) (out chan *ironmq.Message) {
	out = make(chan *ironmq.Message)
	go func() {
		for {
			msg, err := queue.Get()
			log.Println("Poll:", msg, err)
			switch err {
			case ironmq.EmptyQueue:
			case nil:
				out <- msg
			default:
				log.Println("queue error:", err)
			}
			<-time.After(wait)
		}
	}()
	return
}

func main() {
	os.Clearenv()
	if err := env(); err != nil {
		log.Fatal(err)
	}

	client := ironmq.NewClient(
		os.Getenv("IRON_MQ_PROJECT_ID"),
		os.Getenv("IRON_MQ_TOKEN"),
		ironmq.IronAWSUSEast,
	)
	queue := client.Queue("my_queue")

	ch := poll(queue, 10*time.Second) //once every 10 seconds

	go func() {
		<-time.After(5 * time.Second)
		// Put a message on the queue
		id, err := queue.Push("Hello, world!")
		log.Println("Push:", id, err)
	}()

	msg := <-ch
	log.Println(msg)

	// Delete the message
	err := msg.Delete()
	log.Println(err)
}
