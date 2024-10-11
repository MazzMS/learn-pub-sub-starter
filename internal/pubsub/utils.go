package pubsub

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	channel, _, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
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
			msg, err := unmarshaller(delivery.Body)
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
