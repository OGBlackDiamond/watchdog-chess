package engine

import (
	"errors"
	"fmt"
)

func (e *Engine) GenerateLegalMovesForPiece(piece PieceInfo) (uint64, error) {
	pseudoLegalMoves, err := e.GeneratePseudoLegalMovesForPiece(piece)

	if err != nil {
		return uint64(0), err
	}

	legalMoves := uint64(0)

	fromX, fromY, err := MaskToSpace(piece.Mask)

	if err != nil {
		return uint64(0), err
	}

	// loop over all spaces that have a psuedo legal move
	for mask := uint64(1); mask != 0; mask <<= 1 {
		if pseudoLegalMoves&mask == 0 {
			continue
		}

		toX, toY, err := MaskToSpace(mask)

		if err != nil {
			return uint64(0), err
		}

		copy := *e
		_, err = copy.makeMoveUnchecked(fromX, fromY, toX, toY)
		if err != nil {
			continue
		}

		if copy.KingIsChecked(piece.IsWhite) {
			continue
		}

		legalMoves |= mask

	}

	return legalMoves, nil
}

func (e *Engine) GeneratePseudoLegalMovesForPiece(piece PieceInfo) (uint64, error) {

	switch piece.Piece {
	case Pawn:
		return e.GeneratePawnMoves(piece)

	case Rook:
		return e.GenerateLateralMoves(piece)

	case Knight:
		return e.GenerateKnightMoves(piece)

	case Bishop:
		return e.GenerateDiagonalMoves(piece)

	case Queen:
		lateralMask, latErr := e.GenerateLateralMoves(piece)
		diagonalMask, digErr := e.GenerateDiagonalMoves(piece)

		if latErr != nil || digErr != nil {
			return uint64(0), errors.New(latErr.Error() + " " + digErr.Error())
		}

		return lateralMask | diagonalMask, nil

	case King:
		return e.GenerateKingMoves(piece)

	}

	return uint64(0), nil
}

func (e *Engine) GenerateDiagonalMoves(piece PieceInfo) (uint64, error) {

	directions := [][2]int{
		{1, 1},   // northeast
		{-1, 1},  // northwest
		{1, -1},  // southeast
		{-1, -1}, // southwest
	}

	return e.GenerateRangeMoves(piece, directions)
}

func (e *Engine) GenerateLateralMoves(piece PieceInfo) (uint64, error) {

	directions := [][2]int{
		{1, 0},  // east
		{-1, 0}, // west
		{0, 1},  // north
		{0, -1}, // south
	}

	return e.GenerateRangeMoves(piece, directions)
}

func (e *Engine) GenerateKnightMoves(piece PieceInfo) (uint64, error) {

	directions := [][2]int{
		{1, 2},
		{-1, 2},
		{-2, 1},
		{-2, -1},
		{1, -2},
		{-1, -2},
		{2, -1},
		{2, 1},
	}

	return e.GenerateDirectMoves(piece, directions)
}

func (e *Engine) GenerateKingMoves(piece PieceInfo) (uint64, error) {

	directions := [][2]int{
		{0, 1},
		{-1, 1},
		{-1, 0},
		{-1, -1},
		{0, -1},
		{1, -1},
		{1, 0},
		{1, 1},
	}

	// get the mask for the base moves
	moveMask, err := e.GenerateDirectMoves(piece, directions)

	if err != nil {
		return uint64(0), err
	}

	x, y, err := MaskToSpace(piece.Mask)

	if err != nil {
		return uint64(0), err
	}

	// define moveset for castling
	castles := [][2]int{}

	// flip caslting direction if Board is flipped
	sideModifier := -1

	if e.PlayAsWhite {
		sideModifier = 1
	}

	if piece.IsWhite {
		if e.whiteCanCastleQueenSide && e.CastlePathLegal(x, y, 0, piece.IsWhite) {
			castles = append(castles, [2]int{-2 * sideModifier, 0})
		}
		if e.whiteCanCastleKingSide && e.CastlePathLegal(x, y, 7, piece.IsWhite) {
			castles = append(castles, [2]int{2 * sideModifier, 0})
		}
	} else {
		if e.blackCanCastleQueenSide && e.CastlePathLegal(x, y, 0, piece.IsWhite) {
			castles = append(castles, [2]int{-2 * sideModifier, 0})
		}
		if e.blackCanCastleKingSide && e.CastlePathLegal(x, y, 7, piece.IsWhite) {
			castles = append(castles, [2]int{2 * sideModifier, 0})
		}
	}

	castleMask, castleErr := e.GenerateDirectMoves(piece, castles)

	if castleErr != nil {
		return uint64(0), err
	}

	return moveMask | castleMask, nil
}

func (e *Engine) GeneratePawnMoves(piece PieceInfo) (uint64, error) {

	baseDirection := -1

	if (e.PlayAsWhite && !piece.IsWhite) || (!e.PlayAsWhite && piece.IsWhite) {
		baseDirection = 1
	}

	directions := [][2]int{
		{0, baseDirection},
	}

	captures := [][2]int{
		{1, baseDirection},
		{-1, baseDirection},
	}

	x, y, _ := MaskToSpace(piece.Mask)

	canMoveTwice := e.PlayAsWhite && ((piece.IsWhite && y == 6) || (!piece.IsWhite && y == 1)) ||
		!e.PlayAsWhite && ((!piece.IsWhite && y == 6) || (piece.IsWhite && y == 1))

	if canMoveTwice {
		directions = append(directions, [2]int{0, baseDirection * 2})
	}

	// actually start checking and adding to a mask

	mask := uint64(0)

	occupancy := e.Occupancy()

	var enemyOccupancy uint64

	if piece.IsWhite {
		enemyOccupancy = e.BlackOccupancy()
	} else {
		enemyOccupancy = e.WhiteOccupancy()
	}

	// check moves
	for step, dir := range directions {
		if move, err := SpaceToMask(x+dir[0], y+dir[1]); err != nil {
			// in this case we continue and don't return
			// this could be saving us from a wrap-around
			continue
		} else {
			if move&occupancy != 0 {
				continue
			}

			// if the first step was blocked
			if step == 1 {
				if mask == 0 {
					continue
				}
				// if we are allowed to make a second move,
				// set enPassantTarget to the square behind the pawn
				e.enPassantTarget = mask
				e.enPassantPieceMask = move
			}

			mask |= move
		}
	}

	// check captures
	for _, dir := range captures {
		if move, err := SpaceToMask(x+dir[0], y+dir[1]); err != nil {
			continue
		} else {
			if move&enemyOccupancy == 0 && move&e.enPassantTarget == 0 {
				continue
			}

			mask |= move
		}
	}

	return mask, nil
}

func (e *Engine) GenerateDirectMoves(piece PieceInfo, directions [][2]int) (uint64, error) {

	mask := uint64(0)

	x, y, _ := MaskToSpace(piece.Mask)

	var (
		occupancy uint64
	)

	if piece.IsWhite {
		occupancy = e.WhiteOccupancy()
	} else {
		occupancy = e.BlackOccupancy()
	}

	for _, dir := range directions {
		if move, err := SpaceToMask(x+dir[0], y+dir[1]); err != nil {
			continue // this is probably wrap around
		} else {
			if move&occupancy != 0 {
				continue
			}

			mask |= move

		}
	}

	return mask, nil

}

func (e *Engine) GenerateRangeMoves(piece PieceInfo, directions [][2]int) (uint64, error) {

	x, y, err := MaskToSpace(piece.Mask)

	if err != nil {
		fmt.Println("GenerateRangeMoves() failed with: " + err.Error())
		return uint64(0), err
	}

	moves := uint64(0)

	var (
		occupancy      uint64
		enemyOccupancy uint64
	)

	if piece.IsWhite {
		occupancy = e.WhiteOccupancy()
		enemyOccupancy = e.BlackOccupancy()
	} else {
		occupancy = e.BlackOccupancy()
		enemyOccupancy = e.WhiteOccupancy()
	}

	for _, dir := range directions {
		df := dir[0]
		dr := dir[1]

		file := x + df
		rank := y + dr

		for !CheckBounds(file, rank) {
			if mask, err := SpaceToMask(file, rank); err != nil {
				return uint64(0), err
			} else {

				if occupancy&mask != 0 {
					break
				}

				moves |= mask

				if enemyOccupancy&mask != 0 {
					break
				}

				file += df
				rank += dr

			}
		}
	}

	return moves, nil
}
