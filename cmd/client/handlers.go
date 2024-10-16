package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(state routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(state)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetUsername()),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(warRec gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(warRec)

		var msg string

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
			msg = fmt.Sprintf("%s won a war against %s", winner, loser)
		case gamelogic.WarOutcomeDraw:
			msg = fmt.Sprintf("A wat between %s and %s resulted in a draw", winner, loser)
		default:
			fmt.Printf("error: unknown war outcome")
			return pubsub.Ack
		}

		err := pubsub.PublishGob(
			publishCh,
			routing.ExchangePerilTopic,
			routing.GameLogSlug+"."+gs.GetUsername(),
			routing.GameLog{
				CurrentTime: time.Now(),
				Username:    gs.GetUsername(),
				Message:     msg,
			},
		)
		if err != nil {
			log.Printf("error: could not publish log")
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
