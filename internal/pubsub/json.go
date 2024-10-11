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
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	unmarshaller := func(data []byte) (T, error) {
		var msg T
		err := json.Unmarshal(data, &msg)
		if err != nil {
			return msg, err
		}
		return msg, nil
	}
	err := subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		unmarshaller,
	)
	if err != nil {
		return err
	}

	log.Println("Successful JSON subscription")
	return nil
}
