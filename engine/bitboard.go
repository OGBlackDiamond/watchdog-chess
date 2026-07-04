package engine

import (
	"errors"
	"math/bits"
)

func SpaceToMask(x, y int) (uint64, bool) {
	if CheckBounds(x, y) {
		return uint64(0), false
	}

	mask := uint64(1) << ((7-y)*8 + x)

	return mask, true
}

func MaskToSpace(mask uint64) (int, int, error) {

	if mask == 0 {
		return 0, 0, errors.New("mask is empty")
	}

	if mask&(mask-1) != 0 {
		return 0, 0, errors.New("mask has more than one bit set")
	}

	square := bits.TrailingZeros64(mask)

	x := square % 8
	y := 7 - (square / 8)

	return x, y, nil
}

func (e *Engine) WhiteOccupancy() uint64 {
	return e.Board.WhitePieces.Pawns |
		e.Board.WhitePieces.Rooks |
		e.Board.WhitePieces.Knights |
		e.Board.WhitePieces.Bishops |
		e.Board.WhitePieces.Queen |
		e.Board.WhitePieces.King
}

func (e *Engine) BlackOccupancy() uint64 {
	return e.Board.BlackPieces.Pawns |
		e.Board.BlackPieces.Rooks |
		e.Board.BlackPieces.Knights |
		e.Board.BlackPieces.Bishops |
		e.Board.BlackPieces.Queen |
		e.Board.BlackPieces.King
}

func (e *Engine) Occupancy() uint64 {
	return e.WhiteOccupancy() | e.BlackOccupancy()
}

func CheckBounds(x, y int) bool {
	return x > 7 || x < 0 || y > 7 || y < 0
}
