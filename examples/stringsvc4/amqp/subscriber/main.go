package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"

	"github.com/go-kit/kit/endpoint"
	stringsvc "github.com/go-kit/kit/examples/stringsvc4/amqp/svc"
	amqptransport "github.com/go-kit/kit/transport/amqp"
	"github.com/streadway/amqp"
)

func makeUppercaseEndpoint(svc stringsvc.StringService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(stringsvc.UppercaseRequest)
		v, err := svc.Uppercase(ctx, req.S)
		if err != nil {
			return stringsvc.UppercaseResponse{v, err.Error()}, nil
		}
		return stringsvc.UppercaseResponse{v, ""}, nil
	}
}

func makeCountEndpoint(svc stringsvc.StringService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(stringsvc.CountRequest)
		v := svc.Count(ctx, req.S)
		return stringsvc.CountResponse{v}, nil
	}
}

// Transports expose the service to the network. In this fourth example we utilize JSON over AMQP
func main() {
	svc := stringsvc.StringService{}

	amqpURL := flag.String(
		"url",
		"amqp://localhost:5672",
		"URL to AMQP server",
	)

	flag.Parse()

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
	// consume instantiation really depends on the design of the RabbitMQ system
	// for simplicity's sake, I define two separate queues
	// more complex layouts (topic, header exchanges) are also supported
	uppercaseQueue, err := ch.QueueDeclare(
		"stringsvc_uppercase",
		false, // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   //args
	)

	uppercaseMsgs, err := ch.Consume(
		uppercaseQueue.Name,
		"",    // consumer
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)

	countQueue, err := ch.QueueDeclare(
		"stringsvc_count",
		false, // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   //args
	)

	countMsgs, err := ch.Consume(
		countQueue.Name,
		"",    // consumer
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)

	uppercaseAMQPHandler := amqptransport.NewSubscriber(
		makeUppercaseEndpoint(svc),
		decodeUppercaseAMQPRequest,
		amqptransport.EncodeJSONResponse,
	)

	countAMQPHandler := amqptransport.NewSubscriber(
		makeCountEndpoint(svc),
		decodeCountAMQPRequest,
		amqptransport.EncodeJSONResponse,
	)

	uppercaseListener := uppercaseAMQPHandler.ServeDelivery(ch)
	countListener := countAMQPHandler.ServeDelivery(ch)

	forever := make(chan bool)

	go func() {
		for true {
			select {
			case uppercaseDeliv := <-uppercaseMsgs:
				log.Println("received uppercase request")
				uppercaseListener(&uppercaseDeliv)
				uppercaseDeliv.Ack(false) // multiple = false
			case countDeliv := <-countMsgs:
				log.Println("received count request")
				countListener(&countDeliv)
				countDeliv.Ack(false) // multiple = false
			}
		}
	}()

	log.Println("listening")
	<-forever

}

func decodeUppercaseAMQPRequest(ctx context.Context, delivery *amqp.Delivery) (interface{}, error) {
	var request stringsvc.UppercaseRequest
	err := json.Unmarshal(delivery.Body, &request)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func decodeCountAMQPRequest(ctx context.Context, delivery *amqp.Delivery) (interface{}, error) {
	var request stringsvc.CountRequest
	err := json.Unmarshal(delivery.Body, &request)
	if err != nil {
		return nil, err
	}
	return request, nil
}
