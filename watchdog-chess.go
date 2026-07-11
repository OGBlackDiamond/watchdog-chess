package main

import (
	"fmt"

	"github.com/OGBlackDiamond/watchdog-chess/internal/uci"
)

const (
	engineName string = "Watchdog"
	author     string = "OGBlackDiamond - Caden Feller"
)

func main() {

	var input string

	// don't really do anything until we get the uci input
	for input != "uci" {
		fmt.Scanln(&input)
	}

	fmt.Println("id name " + engineName)
	fmt.Println("id author " + author)

	// we can do other setup things here

	// tell the GUI we're ready
	fmt.Println("uciok")

	// this starts UCI talks
	// all processing will happen in threads so this blocking main is fine
	if err := uci.StartUCIHandler(); err != nil {
		fmt.Println("UCI failed with: " + err.Error())
	}

}
