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
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Panicln("Error during queue declaration/binding:", err)
	}

	log.Println("Succssesfully declared and binded a queue")

	gameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState),
	)

infiniteLoop:
	for {
		// get user input
		input := gamelogic.GetInput()

		// no input, go next
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			/*
				The spawn command allows a player to add a new unit to the map under
				their control. Use the gamestate.CommandSpawn method and pass in
				the "words" from the GetInput command. Possible unit types are:

				* infantry
				* cavalry
				* artillery

				Possible locations are:

				* americas
				* europe
				* africa
				* asia
				* antarctica
				* australia
			*/
			err := gameState.CommandSpawn(input)
			if err != nil {
				log.Println(err)
			}

		case "move":
			/*
				The move command allows a player to move their units to a new
				location. It accepts two arguments: the destination, and the ID
				of the unit. Luckily, all that game logic is already implemented
				for you.

				Call the gamestate.CommandMove method and pass in all with
				"words" from the GetInput command. If the move is successful,
				print a message indicating that it worked.
			*/

			// NOTE(molivera): assigned to _ because, at the moment, is not
			// useful, as the command already prints the info
			_, err := gameState.CommandMove(input)
			if err != nil {
				log.Println(err)
			}

		case "status":
			/*
				Use the gamestate.CommandStatus method to print the current
				status of the player's game state.
			*/
			gameState.CommandStatus()

		case "help":
			/*
				Use the gamelogic.PrintClientHelp function to print a list of
				available commands.
			*/
			gamelogic.PrintClientHelp()

		case "spam":
			/*
				For now, just print a message that says
				"Spamming not allowed yet!"
			*/
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			/*
				Use the gamelogic.PrintQuit function to print a message,
				then exit the REPL.
			*/
			gamelogic.PrintQuit()
			break infiniteLoop

		default:
			fmt.Println("Wrong input")
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	log.Println("")
	log.Println("Shutting the client down...")
}

func handlerPause(gamestate *gamelogic.GameState) func(routing.PlayingState) {
	return func(state routing.PlayingState) {
		defer fmt.Print("> ")
		gamestate.HandlePause(state)
	}
}
