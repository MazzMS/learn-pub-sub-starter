package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	log.Println("Starting JSON publication...")

	body, err := json.Marshal(val)
	if err != nil {
		return err
	}

	log.Println("JSON marshal went good")

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return err
	}

	log.Println("JSON published successfully")

	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	channel, _, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
	)
	if err != nil {
		return err
	}

	// get delievery channel
	deliveries, err := channel.Consume(
		queueName,
		"",    // consumer name (auto-generated)
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	// start routine
	go func() {
		for delivery := range deliveries {
			// go through each msg

			// unmarshal it
			var msg T
			err := json.Unmarshal(delivery.Body, &msg)
			if err != nil {
				delivery.Nack(false, true) // no acknowledge
				continue
			}

			// do whatever is supposed
			acktype := handler(msg)

			switch acktype {
			case Ack:
				err = delivery.Ack(false)
				if err != nil {
					log.Printf("Error acknowledging message: %v\n", err)
				}
				log.Println("Acknowledging message")
			case NackRequeue:
				err = delivery.Nack(false, true)
				if err != nil {
					log.Printf("Error not acknowledging and requeueing message: %v\n", err)
				}
				log.Println("Not acknowledging message, requeueing")
			case NackDiscard:
				err = delivery.Nack(false, false)
				if err != nil {
					log.Printf("Error not acknowledging and discarding message: %v\n", err)
				}
				log.Println("Not acknowledging message, discarding")
			}
		}
	}()

	return nil
}
