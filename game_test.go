package main

import "testing"

func TestScreenBoardCoordinateConversion(t *testing.T) {
	tests := []struct {
		name        string
		playWhite   bool
		screenX     int
		screenY     int
		wantBoardX  int
		wantBoardY  int
		wantScreenX int
		wantScreenY int
	}{
		{
			name:        "white perspective identity",
			playWhite:   true,
			screenX:     4,
			screenY:     7,
			wantBoardX:  4,
			wantBoardY:  7,
			wantScreenX: 4,
			wantScreenY: 7,
		},
		{
			name:        "black perspective flips both axes",
			playWhite:   false,
			screenX:     3,
			screenY:     0,
			wantBoardX:  4,
			wantBoardY:  7,
			wantScreenX: 3,
			wantScreenY: 0,
		},
	}

	oldPlayAsWhite := playAsWhite
	t.Cleanup(func() { playAsWhite = oldPlayAsWhite })

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			playAsWhite = tt.playWhite

			boardX, boardY := screenToBoard(tt.screenX, tt.screenY)
			if boardX != tt.wantBoardX || boardY != tt.wantBoardY {
				t.Fatalf("screenToBoard(%d, %d) = (%d, %d), want (%d, %d)", tt.screenX, tt.screenY, boardX, boardY, tt.wantBoardX, tt.wantBoardY)
			}

			screenX, screenY := boardToScreen(boardX, boardY)
			if screenX != tt.wantScreenX || screenY != tt.wantScreenY {
				t.Fatalf("boardToScreen(%d, %d) = (%d, %d), want (%d, %d)", boardX, boardY, screenX, screenY, tt.wantScreenX, tt.wantScreenY)
			}
		})
	}
}
