package main

import (
	"fmt"

	chessengine "github.com/OGBlackDiamond/watchdog-chess/engine"
	"github.com/hajimehoshi/ebiten/v2"
)

/**
* This file will pretty much just clean up and handle
* player interaction and other stuff like that
 */

var (
	dragX, dragY    int
	isDragging      bool = false
	clickLegalMoves []chessengine.Move

	playAsWhite = false

	whiteToMove bool = true
)

func handleLeftPress() error {
	screenX, screenY := ebiten.CursorPosition()

	screenX /= int(tileSize)
	screenY /= int(tileSize)
	dragX, dragY = screenToBoard(screenX, screenY)

	fmt.Printf("screen: %d, %d board: %d, %d\n", screenX, screenY, dragX, dragY)

	pieceInfo, err := engine.GetBitBoardForSquare(dragX, dragY)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	clickLegalMoves = make([]chessengine.Move, 0, 64)

	if err := engine.GenerateLegalMovesForPiece(*pieceInfo, &clickLegalMoves); err != nil {
		fmt.Println(err.Error())
		return err
	}

	graphics.DrawPieceOnCursor(*pieceInfo)
	isDragging = true

	return nil
}

func handleLeftRelease() error {

	isDragging = false

	screenX, screenY := ebiten.CursorPosition()

	screenX /= int(tileSize)
	screenY /= int(tileSize)
	x, y := screenToBoard(screenX, screenY)

	_, err := engine.MakeMove(chessengine.Move{FromX: dragX, FromY: dragY, ToX: x, ToY: y})

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	// a move was actually made if we make it here
	whiteToMove = !whiteToMove

	return nil
}

func screenToBoard(x, y int) (int, int) {
	if playAsWhite {
		return x, y
	}

	return 7 - x, 7 - y
}

func boardToScreen(x, y int) (int, int) {
	if playAsWhite {
		return x, y
	}

	return 7 - x, 7 - y
}
