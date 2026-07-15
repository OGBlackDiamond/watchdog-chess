// Package uci handles all communication between the engine and a GUI using the UCI protocol
package uci

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/OGBlackDiamond/watchdog-chess/internal/engine"
)

var (
	e engine.Engine
)

// StartUCIHandler handles all general UCI command throughput
func StartUCIHandler() error {

	numThreads := runtime.NumCPU()

	fmt.Printf("option name Threads type spin default %d min 1 max %d\n", numThreads, numThreads)
	fmt.Printf("option name Hash type spin default 64 min 1 max 1024\n")

	// we can do other setup things here

	// tell the GUI we're ready
	fmt.Println("uciok")

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
			e.SetTTSize(64)

		case "go":
			// start the engine searching
			bestMove, _, err := e.ChooseMove(10, numThreads)
			if err != nil {
				return err
			}
			fmt.Printf("bestmove %s\n", bestMove.ToAlgNot())

		case "stop":
			// stop the engine searching

		case "position":
			// set the position in the engine; a bad position command should
			// not kill the engine, just report and keep listening
			if err := e.SetEnginePosition(args); err != nil {
				fmt.Printf("info string %s\n", err.Error())
			}

		case "ponderhit":
			// player played the expected position

		}

	}

	return nil
}

func newGame() error {
	e.SetTTSize(64)
	return nil
}
