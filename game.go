package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

/**
* This file will pretty much just clean up and handle
* player interaction and other stuff like that
 */

var (
	dragX, dragY int
	isDragging bool = false
	clickLegalMoves uint64
)

func handleLeftPress() error {
	dragX, dragY = ebiten.CursorPosition()

	dragX /= int(tileSize)
	dragY /= int(tileSize)

	fmt.Printf("%d, %d\n", dragX, dragY)

	pieceInfo, bbErr := engine.GetBitBoardForSquare(dragX, dragY)

	if bbErr != nil {
		fmt.Println(bbErr.Error())
		return bbErr
	}

	legalMoves, legalMovesErr := engine.GenerateLegalMoves(*pieceInfo)

	if legalMovesErr != nil {
		fmt.Println(legalMovesErr.Error())
		return legalMovesErr
	}

	clickLegalMoves = legalMoves

	graphics.DrawPieceOnCursor(*pieceInfo)
	isDragging = true

	return nil
}

func handleLeftRelease() error {
	
	isDragging = false

	x, y := ebiten.CursorPosition()

	x /= int(tileSize)
	y /= int(tileSize)

	_, err := engine.MakeMove(dragX, dragY, x, y)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}
