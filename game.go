package main

import (
	"fmt"
	"strings"

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

	clickedPiece chessengine.PieceInfo

	playAsWhite = true

	whiteToMove bool = true
)

func handleLeftPress() error {
	screenX, screenY := ebiten.CursorPosition()

	screenX /= int(tileSize)
	screenY /= int(tileSize)
	dragX, dragY = screenToBoard(screenX, screenY)

	fmt.Printf("screen: %d, %d board: %d, %d\n", screenX, screenY, dragX, dragY)

	pieceInfo, spaceIsOccupied := engine.GetBitBoardForSquare(dragX, dragY)

	if !spaceIsOccupied {
		fmt.Println("Space is empty")
		return nil
	}

	clickedPiece = pieceInfo

	occ := chessengine.OccupancyInfo{
		White: engine.WhiteOccupancy(),
		Black: engine.BlackOccupancy(),
		All: engine.Occupancy(),
	}

	clickLegalMoves = make([]chessengine.Move, 0, 64)

	if err := engine.GenerateLegalMovesForPiece(pieceInfo, &clickLegalMoves, occ); err != nil {
		fmt.Println(err.Error())
		return err
	}

	graphics.DrawPieceOnCursor(pieceInfo)
	isDragging = true

	return nil
}

func handleLeftRelease() error {

	isDragging = false

	screenX, screenY := ebiten.CursorPosition()

	screenX /= int(tileSize)
	screenY /= int(tileSize)
	x, y := screenToBoard(screenX, screenY)

	promotion := chessengine.NONE

	if screenY == 0 && clickedPiece.Piece == chessengine.Pawn {
		var char string

		fmt.Print("What would you like to promote to? (q,r,b,k) :: ")
		fmt.Scan(&char)

		char = strings.ToLower(char)
		char = char[:1]

		switch char {
		case "q":
			promotion = chessengine.Queen
		case "r":
			promotion = chessengine.Rook
		case "b":
			promotion = chessengine.Bishop
		case "k":
			promotion = chessengine.Knight
		}
	}

	moveMade := chessengine.Move{
		FromX: dragX,
		FromY: dragY,
		ToX: x,
		ToY: y,
		Promotion: promotion,
	}


	didMakeMove, err := engine.MakeMove(moveMade)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if didMakeMove {
		lastMoveMade = moveMade
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
