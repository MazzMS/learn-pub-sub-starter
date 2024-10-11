package main

import (
	"fmt"
	"log"
	"strconv"

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
		log.Fatalln("Error during connection:", err)
	}
	defer connection.Close()
	defer log.Println("Closing connection...")

	log.Println("Connection established!")

	// publish channel creation
	publishCh, err := connection.Channel()
	if err != nil {
		log.Fatalln("Error during channel creation:", err)
	}

	// get username
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalln("Error during connection:", err)
	}

	// generate game state
	gs := gamelogic.NewGameState(username)

	// subscribe to 'war' queue
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable,
		handlerWar(gs, publishCh),
	)
	if err != nil {
		log.Fatalln("Error during subscription to pause:", err)
	}
	log.Printf("Succssesfully subscribed '%s' queue", routing.WarRecognitionsPrefix)

	// subscribe to 'pause' queue
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, gs.GetUsername()),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalln("Error during subscription to pause:", err)
	}
	log.Printf("Succssesfully subscribed '%s' queue", routing.PauseKey)

	// subscribe to 'army moves' queue
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, gs.GetUsername()),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.Transient,
		handlerMove(gs, publishCh),
	)
	if err != nil {
		log.Fatalln("Error during army moves subscription:", err)
	}

	log.Printf("Succssesfully subscribed '%s' queue", routing.ArmyMovesPrefix)

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
			err := gs.CommandSpawn(input)
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
			move, err := gs.CommandMove(input)
			if err != nil {
				log.Println(err)
			}

			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, move.Player.Username),
				move,
			)
			if err != nil {
				log.Printf("There was an error publishing the move, it will not be applied: %v\n", err)
				for _, unit := range move.Units {
					id := strconv.Itoa(unit.ID)
					location := fmt.Sprint(unit.Location)

					_, err := gs.CommandMove([]string{"move", location, id})
					if err != nil {
						log.Println(err)
					}
				}
			}
			fmt.Printf("Moved %d units to %s\n", len(move.Units), move.ToLocation)

		case "status":
			/*
				Use the gamestate.CommandStatus method to print the current
				status of the player's game state.
			*/
			gs.CommandStatus()

		case "help":
			/*
				Use the gamelogic.PrintClientHelp function to print a list of
				available commands.
			*/
			gamelogic.PrintClientHelp()

		case "spam":
			secondWord := input[1]
			times, err := strconv.Atoi(secondWord)
			if err != nil {
				fmt.Println("bad number")
			}

			for i := 0; i < times; i++ {
				malLog := gamelogic.GetMaliciousLog()
				pubsub.PublishGob(
					publishCh,
					routing.ExchangePerilTopic,
					"game_logs."+gs.GetUsername(),
					malLog,
				)
			}

		case "quit":
			/*
				Use the gamelogic.PrintQuit function to print a message,
				then exit the REPL.
			*/
			gamelogic.PrintQuit()
			return

		default:
			fmt.Println("Wrong input")
		}
	}
}
