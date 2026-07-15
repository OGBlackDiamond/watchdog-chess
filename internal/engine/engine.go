// Package engine handles the search and descision making side of the computer
package engine

import (
	"errors"
	"strings"

	"github.com/OGBlackDiamond/watchdog-chess/internal/board"
	"github.com/OGBlackDiamond/watchdog-chess/internal/fen"
)

type Engine struct {
	b             board.Board
	positionSet   bool
	lastPosStrLen int

	tt TT
}

func (e *Engine) SetTTSize(mb int) {
	e.tt.Resize(mb)
}

func (e *Engine) SetEnginePosition(args []string) error {
	if len(args) == 0 {
		return errors.New("invalid engine position: missing position type")
	}

	if e.positionSet && len(args) == e.lastPosStrLen+1 {
		// the position input is a continuation

		// make the most recent move to update board state
		if err := e.b.MakeMoveFromAlgNot(args[len(args)-1]); err != nil {
			return errors.New("setEnginePosition failed with: " + err.Error())
		}

		e.lastPosStrLen++

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

	e.b = *bState

	// skip the optional "moves" keyword before the move list
	if tokenIndex < len(args) && args[tokenIndex] == "moves" {
		tokenIndex++
	}

	for _, moveStr := range args[tokenIndex:] {
		if err := e.b.MakeMoveFromAlgNot(moveStr); err != nil {
			return errors.New("setEnginePosition failed with: " + err.Error())
		}
	}

	e.lastPosStrLen = len(args)
	e.positionSet = true

	return nil
}
