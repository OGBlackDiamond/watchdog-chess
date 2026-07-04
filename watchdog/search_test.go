package watchdog

import (
	"testing"

	"github.com/OGBlackDiamond/watchdog-chess/engine"
)

func TestChooseMoveDepthThreeInitialPosition(t *testing.T) {
	tests := []struct {
		name        string
		playAsWhite bool
		whiteToMove bool
	}{
		{name: "playing white", playAsWhite: true, whiteToMove: false},
		{name: "playing black", playAsWhite: false, whiteToMove: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := engine.NewEngine(tt.playAsWhite)

			move, found, err := ChooseMove(e, 3, tt.whiteToMove)
			if err != nil {
				t.Fatalf("ChooseMove returned error: %v", err)
			}
			if !found {
				t.Fatal("ChooseMove did not find a move")
			}

			if _, err := e.MakeMoveUnchecked(move); err != nil {
				t.Fatalf("ChooseMove returned illegal move %+v: %v", move, err)
			}
		})
	}
}
