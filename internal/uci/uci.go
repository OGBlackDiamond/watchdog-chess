// Package uci handles all communication between the engine and a GUI using the UCI protocol
package uci

import (
	"errors"
	"fmt"
	"strings"

	"github.com/OGBlackDiamond/watchdog-chess/internal/board"
	"github.com/OGBlackDiamond/watchdog-chess/internal/fen"
)

var (

	boardState board.Board

	// this stores the length of the last position string
	lastPosStrLen int = -1

)

// StartUCIHandler handles all general UCI command throughput
func StartUCIHandler() {

	for {

		var input string

		// all of the searching and processing will happen in threads so this blocking is fine
		fmt.Scanln(&input)

		var firstToken string
		spaceIndex := strings.Index(input, " ")

		if spaceIndex != -1 {
			firstToken = input[:spaceIndex]
		} else {
			firstToken = input
		}


		switch firstToken {
		case "quit":
			// probably do something else here to stop the rest of the program
			return
			
		case "isready":
			// we can do some checks here at some point
			fmt.Println("readyok")

		case "ucinewgame":
			// probably do engine init here
			newGame()

		case "setoption":
			// if we allow config, set it up
		
		case "go":
			// start the engine searching
			fmt.Println("bestmove xyx1 ponder abc2")

		case "stop":
			// stop the engine searching
		
		case "position":
			// set the position in the engine
			setEnginePosition(input[spaceIndex + 1:])

		case "ponderhit":
			// player played the expected position


		}

	}
}


func newGame() error {

	return nil
}

func setEnginePosition(command string) error {

	tokens := strings.Split(strings.TrimSpace(command), " ")

	if len(tokens) == lastPosStrLen + 1 {
		// the position input is a continuation
		
		// make the most recent move to update board state
		if err := boardState.MakeMoveFromAlgNot(tokens[len(tokens) - 1]); err != nil {
			return errors.New("setEnginePosition failed with: " + err.Error())
		}

		lastPosStrLen++

		return nil
	}


	var fenString string

	tokenIndex := 0

	switch tokens[tokenIndex] {
	case "startpos":
		fenString = fen.StartingPositionFEN
	case "fen":
		tokenIndex++
		fenString = tokens[tokenIndex]
	default:
		return errors.New("invalid engine position")
	}


	bState, err := fen.NewBoardFromFen(fenString)

	if err != nil {
		return errors.New("setEnginePosition failed with: " + err.Error())
	}

	boardState = *bState

	tokenIndex++

	for _, moveStr := range tokens[tokenIndex:] {
		if err := boardState.MakeMoveFromAlgNot(moveStr); err != nil {
			return errors.New("setEnginePosition failed with: " + err.Error())
		}
	}

	lastPosStrLen = len(tokens)

	return nil
}
