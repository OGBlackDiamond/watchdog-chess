package board

import "math/bits"

// IsSquareAttacked checks if the provided square is attacked using super-piece checks
func (b *Board) IsSquareAttacked(sq int, byWhite bool) bool {
	side := blackIndex
	colorMask := 0b0000
	if byWhite {
		side = whiteIndex
		colorMask = 0b1000
	}

	if pawnAttackers[side][sq]&b.Bitboards[Pawn+Piece(colorMask)] != 0 {
		return true
	}

	if knightAttacks[sq]&b.Bitboards[Knight+Piece(colorMask)] != 0 {
		return true
	}

	if kingAttacks[sq]&b.Bitboards[King+Piece(colorMask)] != 0 {
		return true
	}

	if bishopAttacks(sq, b.Occupancy)&(b.Bitboards[Bishop+Piece(colorMask)]|b.Bitboards[Queen+Piece(colorMask)]) != 0 {
		return true
	}

	if rookAttacks(sq, b.Occupancy)&(b.Bitboards[Rook+Piece(colorMask)]|b.Bitboards[Queen+Piece(colorMask)]) != 0 {
		return true
	}

	return false

}

func (b *Board) KingIsChecked(isWhite bool) bool {
	kingBB := b.Bitboards[King]
	if isWhite {
		kingBB = b.Bitboards[King+0b1000]
	}

	// a missing king (only possible from an illegal input position, since the
	// search verifies pseudo-legal moves) counts as in check, so the search
	// treats king capture as mate instead of indexing out of bounds
	if kingBB == 0 {
		return true
	}

	kingSq := bits.TrailingZeros64(kingBB)
	return b.IsSquareAttacked(kingSq, !isWhite)
}

func (b *Board) CanCastleKing(home int, kingSide bool, isWhite bool) bool {

	if kingSide {
		kingSideBetween := (uint64(1) << (home + 1)) | (uint64(1) << (home + 2))
		return kingSideBetween&b.Occupancy == 0 &&
			!b.IsSquareAttacked(home+1, !isWhite) &&
			!b.IsSquareAttacked(home+2, !isWhite)
	}

	return false
}

func (b *Board) CanCastleQueen(home int, queenSide bool, isWhite bool) bool {

	if queenSide {
		queenSideBetween := (uint64(1) << (home - 1)) | (uint64(1) << (home - 2)) | (uint64(1) << (home - 3))
		return queenSideBetween&b.Occupancy == 0 &&
			!b.IsSquareAttacked(home-1, !isWhite) &&
			!b.IsSquareAttacked(home-2, !isWhite)
	}

	return false
}
