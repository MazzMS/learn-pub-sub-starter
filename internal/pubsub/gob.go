package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	log.Println("Starting Gob publication...")

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(val)
	if err != nil {
		return err
	}

	log.Println("Gob encoding went good")

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        buffer.Bytes(),
		},
	)
	if err != nil {
		return err
	}

	log.Println("Gob published successfully")

	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	unmarshaller := func(data []byte) (T, error) {
		buffer := bytes.NewBuffer(data)
		decoder := gob.NewDecoder(buffer)
		var msg T
		err := decoder.Decode(&msg)
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

	log.Println("Successful Gob subscription")
	return nil
}
