package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gamestate *gamelogic.GameState) func(routing.PlayingState) {
	return func(state routing.PlayingState) {
		defer fmt.Print("> ")
		gamestate.HandlePause(state)
	}
}

func handlerMove(gamestate *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(state gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		gamestate.HandleMove(state)
	}
}
