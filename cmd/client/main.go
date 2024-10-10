package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	connectionUrl := "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connectionUrl)
	if err != nil {
		log.Panicln("Error during connection:", err)
	}
	defer connection.Close()
	defer log.Println("Closing connection")

	log.Println("Connection established!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Panicln("Error during connection:", err)
	}

	// channel, queue, err := pubsub.DeclareAndBind(
	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s",routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Panicln("Error during queue declaration/binding:", err)
	}


	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	log.Println("")
	log.Println("Shutting the client down...")
}
