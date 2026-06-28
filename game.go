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
	bitmapInEffect *uint64
	clickMask uint64
)

func handleLeftPress() error {
	dragX, dragY = ebiten.CursorPosition()

	fmt.Printf("%d, %d\n", dragX/int(tileSize), dragX/int(tileSize))

	dragX /= int(tileSize)
	dragY /= int(tileSize)

	pieceInfo, err := engine.GetBitBoardForSquare(dragX, dragY)

	if err != nil {
		// uh idk
		return err
	}
	if pieceInfo.bitboard == nil {
		return err
	}

	bitmapInEffect = pieceInfo.bitboard
	clickMask = pieceInfo.mask

	*bitmapInEffect ^= clickMask

	graphics.DrawPieceOnCursor(*pieceInfo)

	return nil
}

func handleLeftRelease() error {
	
	graphics.StopDrawingPieceOnCursor()

	x, y := ebiten.CursorPosition()

	x /= int(tileSize)
	y /= int(tileSize)

	makeMask := uint64(1) << ((7-y) * 8 + x)

	if x != dragX || y != dragY {
		if bitmapInEffect == nil {
			return nil
		}

		*bitmapInEffect ^= makeMask
	}

	return nil
}
