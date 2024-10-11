package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	// vars initialization
	channel := &amqp.Channel{}
	queue := amqp.Queue{}

	// checking that enum is in bounds
	if simpleQueueType < 0 || simpleQueueType > 1 {
		return channel, queue, fmt.Errorf("QueueType does not has a proper value: %d", simpleQueueType)
	}

	// channel creation
	channel, err := conn.Channel()
	if err != nil {
		return channel, queue, err
	}

	// queue creation
	queue, err = channel.QueueDeclare(
		queueName,
		simpleQueueType == Durable,
		simpleQueueType == Transient,
		simpleQueueType == Transient,
		false,
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
	)
	if err != nil {
		return channel, queue, err
	}

	// Bind queue
	err = channel.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return channel, queue, err
	}

	// return values
	return channel, queue, nil
}
