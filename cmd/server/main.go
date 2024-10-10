package main

import (
	"log"
	"os"
	"os/signal"

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
		log.Panicln("Error during connection:", err)
	}
	defer connection.Close()
	defer log.Println("Closing connection")

	log.Println("Connection established!")

	channel, err := connection.Channel()
	if err != nil {
		log.Panicln("Error during channel creation:", err)
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
		log.Panicln("Error during channel creation:", err)
	}

	// Print what users can do
	gamelogic.PrintServerHelp()

	infiniteLoop:
	for {
		// get user input
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
				log.Panicln("Error while sending pause message:", err)
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
				log.Panicln("Error while sending resume message:", err)
			}
			log.Println("Resume message sent")

		case "quit":
			log.Println("Quitting")
			break infiniteLoop

		default:
			log.Printf("Command undefined: %s\n", possibleInputs[0])
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	log.Println("")
	log.Println("Shutting the server down...")
}
