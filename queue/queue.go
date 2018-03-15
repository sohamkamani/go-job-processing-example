package queue

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

var conn *amqp.Connection

func Init(c string) {
	var err error
	// Initialize the package level "conn" variable that represents the connection the the rabbitmq server
	conn, err = amqp.Dial(c)
	if err != nil {
		log.Fatalf("could not connect to rabbitmq: %v", err)
		panic(err)
	}
}

func Publish(q string, msg []byte) error {
	// create a channel through which we publish
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// create the payload with the message that we specify in the arguments
	payload := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         msg,
	}

	// publish the message to the queue specified in the arguments
	if err := ch.Publish("", q, false, false, payload); err != nil {
		return fmt.Errorf("[Publish] failed to publish to queue %v", err)
	}

	return nil
}

func Subscribe(qName string) (<-chan amqp.Delivery, func(), error) {
	// create a channel through which we publish
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}
	// assert that the queue exists (creates a queue if it doesn't)
	q, err := ch.QueueDeclare(qName, false, false, false, false, nil)

	// create a channel in go, through which incoming messages will be received
	c, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	// return the created channel
	return c, func() { ch.Close() }, err
}
