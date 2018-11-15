package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"sync"

	stringsvc "github.com/go-kit/kit/examples/stringsvc4/amqp/svc"
	amqptransport "github.com/go-kit/kit/transport/amqp"
	"github.com/streadway/amqp"
)

func main() {
	amqpURL := flag.String(
		"url",
		"amqp://localhost:5672",
		"URL to AMQP server",
	)

	word := flag.String(
		"word",
		"bird is the word",
		"Word to input to stringsvc",
	)

	// connect to AMQP
	conn, err := amqp.Dial(*amqpURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	replyQueue, err := ch.QueueDeclare(
		"stringsvc_publisher",
		false, // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   //args
	)

	uppercasePublisher := amqptransport.NewPublisher(
		ch,
		&replyQueue,
		encodeUppercaseAMQPRequest,
		decodeUppercaseAMQPResponse,
		amqptransport.PublisherBefore(
			// queue name specified by subscriber
			amqptransport.SetPublishKey("stringsvc_uppercase"),
		),
	)
	uppercaseEndpoint := uppercasePublisher.Endpoint()

	countPublisher := amqptransport.NewPublisher(
		ch,
		&replyQueue,
		encodeCountAMQPRequest,
		decodeCountAMQPResponse,
		amqptransport.PublisherBefore(
			// queue name specified by subscriber
			amqptransport.SetPublishKey("stringsvc_count"),
		),
	)
	countEndpoint := countPublisher.Endpoint()

	var wg sync.WaitGroup
	uppercaseRequest := stringsvc.UppercaseRequest{S: *word}
	wg.Add(1)
	go func() {
		log.Println("sending uppercase request")
		defer wg.Done()
		response, err := uppercaseEndpoint(context.Background(), uppercaseRequest)
		if err != nil {
			log.Println("error with uppercase request", err)
		} else {
			log.Println("uppercase response: ", response)
		}
	}()
	countRequest := stringsvc.CountRequest{S: *word}
	wg.Add(1)
	go func() {
		log.Println("sending count request")
		defer wg.Done()
		response, err := countEndpoint(context.Background(), countRequest)
		if err != nil {
			log.Println("error with count request", err)
		} else {
			log.Println("count response: ", response)
		}
		wg.Done()
	}()
	wg.Wait()
	log.Println("done")
}

var badRequest = errors.New("bad request type")

func encodeUppercaseAMQPRequest(ctx context.Context, publishing *amqp.Publishing, request interface{}) error {
	uppercaseRequest, ok := request.(stringsvc.UppercaseRequest)
	if !ok {
		return badRequest
	}
	b, err := json.Marshal(uppercaseRequest)
	if err != nil {
		return err
	}
	publishing.Body = b
	return nil
}

func encodeCountAMQPRequest(ctx context.Context, publishing *amqp.Publishing, request interface{}) error {
	countRequest, ok := request.(stringsvc.CountRequest)
	if !ok {
		return badRequest
	}
	b, err := json.Marshal(countRequest)
	if err != nil {
		return err
	}
	publishing.Body = b
	return nil
}

func decodeUppercaseAMQPResponse(ctx context.Context, delivery *amqp.Delivery) (interface{}, error) {
	var response stringsvc.UppercaseResponse
	err := json.Unmarshal(delivery.Body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func decodeCountAMQPResponse(ctx context.Context, delivery *amqp.Delivery) (interface{}, error) {
	var response stringsvc.CountResponse
	err := json.Unmarshal(delivery.Body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
