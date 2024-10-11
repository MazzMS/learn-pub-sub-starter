package main

import (
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	log.Println("Starting Peril server")

	connectionUrl := "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connectionUrl)
	if err != nil {
		log.Fatalln("Error during connection:", err)
	}
	defer connection.Close()
	defer log.Println("Closing connection")

	log.Println("Connection established!")

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalln("Error during channel creation:", err)
	}

	// channel, queue, err := pubsub.DeclareAndBind(
	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		"game_logs",
		"game_logs.*",
		pubsub.Durable,
	)
	if err != nil {
		log.Fatalln("Error during queue declaration/binding:", err)
	}

	err = pubsub.PublishJSON(
		channel,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{
			IsPaused: true,
		},
	)
	if err != nil {
		log.Fatalln("Error during channel creation:", err)
	}

	// Subscribe to logs
	err = pubsub.SubscribeGob(
		connection,
		routing.ExchangePerilTopic,
		"game_logs",
		"game_logs.*",
		pubsub.Durable,
		handlerLogs(),
	)
	if err != nil {
		log.Fatalln("error during logs subscription (gob):", err)
	}

	// Print what users can do
	gamelogic.PrintServerHelp()

	for {
		// get input
		possibleInputs := gamelogic.GetInput()

		// no input, go next
		if len(possibleInputs) == 0 {
			continue
		}

		switch possibleInputs[0] {
		case "pause":
			log.Println("Sending a pause message...")
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				log.Fatalln("Error while sending pause message:", err)
			}
			log.Println("Pause message sent")

		case "resume":
			log.Println("Sending a resume message...")
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				log.Fatalln("Error while sending resume message:", err)
			}
			log.Println("Resume message sent")

		case "quit":
			log.Println("Quitting")
		return

		default:
			log.Printf("Command undefined: %s\n", possibleInputs[0])
		}
	}
}
