// Package uci handles all communication between the engine and a GUI using the UCI protocol
package uci

import (
	"bufio"
	"errors"
	"fmt"
	"os"
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
func StartUCIHandler() error {

	scanner := bufio.NewScanner(os.Stdin)

	// all of the searching and processing will happen in threads so this blocking is fine
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)

		if len(fields) == 0 {
			continue
		}

		command := fields[0]
		args := fields[1:]

		switch command {
		case "quit":
			// probably do something else here to stop the rest of the program
			return nil

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
			//fmt.Println("bestmove d7d5")
			randMove, err := boardState.GenerateLegalMovesForPosition()
			if err != nil {
				return err
			}
			fmt.Printf("bestmove %s\n", randMove[0].ToAlgNot())

		case "stop":
			// stop the engine searching

		case "position":
			// set the position in the engine
			if err := setEnginePosition(args); err != nil {
				return nil
			}

		case "ponderhit":
			// player played the expected position

		}

	}

	return nil
}

func newGame() error {

	return nil
}

func setEnginePosition(args []string) error {

	if len(args) == lastPosStrLen+1 {
		// the position input is a continuation

		// make the most recent move to update board state
		if err := boardState.MakeMoveFromAlgNot(args[len(args)-1]); err != nil {
			return errors.New("setEnginePosition failed with: " + err.Error())
		}

		lastPosStrLen++

		return nil
	}

	var fenString string

	tokenIndex := 0

	switch args[tokenIndex] {
	case "startpos":
		fenString = fen.StartingPositionFEN
	case "fen":
		// the fen string contains 6 'tokens'
		for tokenIndex < 7 {
			tokenIndex++
			fenString += args[tokenIndex] + " "
		}
	default:
		return errors.New("invalid engine position")
	}

	bState, err := fen.NewBoardFromFen(fenString)

	if err != nil {
		return errors.New("setEnginePosition failed with: " + err.Error())
	}

	boardState = *bState

	tokenIndex++

	for _, moveStr := range args[tokenIndex:] {
		if err := boardState.MakeMoveFromAlgNot(moveStr); err != nil {
			return errors.New("setEnginePosition failed with: " + err.Error())
		}
	}

	lastPosStrLen = len(args)

	return nil
}
