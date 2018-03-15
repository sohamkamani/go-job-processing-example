package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sohamkamani/go-job-processing-example/queue"
)

func main() {
	// initialize the queue connection
	queue.Init("amqp://localhost")
	// start a publisher or worker depending on the command line argument
	if os.Args[1] == "worker" {
		worker()
	} else {
		publisher()
	}
}

func publisher() {
	// the publisher publishes the message "1,1" every 500 milliseconds, perpetually
	for {
		if err := queue.Publish("add_q", []byte("1,1")); err != nil {
			panic(err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func worker() {
	// obtain the channel which we subscribe to
	msgs, close, err := queue.Subscribe("add_q")
	if err != nil {
		panic(err)
	}
	defer close()
	forever := make(chan bool)

	go func() {
		// Receive messages from the channel forever
		for d := range msgs {
			// everytime a message is received, convert it to numbers, add the numbers
			i1, i2 := toNums(d.Body)
			// then print the result to STDOUT, along with the time
			fmt.Println(time.Now().Format("01-02-2006 15:04:05"), "::", i1+i2)
			// acknowledge the message so that it is cleared from the queue
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func toNums(b []byte) (int, int) {
	s := string(b)
	ss := strings.Split(s, ",")
	i1, _ := strconv.Atoi(ss[0])
	i2, _ := strconv.Atoi(ss[1])
	return i1, i2
}
