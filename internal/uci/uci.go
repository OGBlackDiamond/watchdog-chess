// Package uci handles all communication between the engine and a GUI using the UCI protocol
package uci

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/OGBlackDiamond/watchdog-chess/internal/board"
	"github.com/OGBlackDiamond/watchdog-chess/internal/engine"
	"github.com/OGBlackDiamond/watchdog-chess/internal/fen"
)

var (
	boardState board.Board

	// this stores the length of the last position string
	lastPosStrLen int = -1
)

// StartUCIHandler handles all general UCI command throughput
func StartUCIHandler() error {

	numCoresToUse := runtime.NumCPU()

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
			bestMove, _, err := engine.ChooseMove(&boardState, 9, numCoresToUse)
			if err != nil {
				return err
			}
			fmt.Printf("bestmove %s\n", bestMove.ToAlgNot())

		case "stop":
			// stop the engine searching

		case "position":
			// set the position in the engine; a bad position command should
			// not kill the engine, just report and keep listening
			if err := setEnginePosition(args); err != nil {
				fmt.Printf("info string %s\n", err.Error())
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
	if len(args) == 0 {
		return errors.New("invalid engine position: missing position type")
	}

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
		tokenIndex = 1
	case "fen":
		// the fen string contains 6 'tokens' following the keyword
		if len(args) < 7 {
			return errors.New("invalid engine position: fen requires 6 fields")
		}
		fenString = strings.Join(args[1:7], " ")
		tokenIndex = 7
	default:
		return errors.New("invalid engine position")
	}

	bState, err := fen.NewBoardFromFen(fenString)

	if err != nil {
		return errors.New("setEnginePosition failed with: " + err.Error())
	}

	boardState = *bState

	// skip the optional "moves" keyword before the move list
	if tokenIndex < len(args) && args[tokenIndex] == "moves" {
		tokenIndex++
	}

	for _, moveStr := range args[tokenIndex:] {
		if err := boardState.MakeMoveFromAlgNot(moveStr); err != nil {
			return errors.New("setEnginePosition failed with: " + err.Error())
		}
	}

	lastPosStrLen = len(args)

	return nil
}
