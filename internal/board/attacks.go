package board

import "math/bits"

// IsSquareAttacked checks if the provided square is attacked using super-piece checks
func (b *Board) IsSquareAttacked(sq int, byWhite bool) bool {
	attackers := b.BlackPieces
	side := blackIndex
	if byWhite {
		attackers = b.WhitePieces
		side = whiteIndex
	}

	if pawnAttackers[side][sq]&attackers.Pawns != 0 {
		return true
	}

	if knightAttacks[sq]&attackers.Knights != 0 {
		return true
	}

	if kingAttacks[sq]&attackers.King != 0 {
		return true
	}

	if bishopAttacks(sq, b.Occupancy)&(attackers.Bishops|attackers.Queen) != 0 {
		return true
	}

	if rookAttacks(sq, b.Occupancy)&(attackers.Rooks|attackers.Queen) != 0 {
		return true
	}

	return false

}

func (b *Board) KingIsChecked(isWhite bool) bool {
	kingBB := b.BlackPieces.King
	if isWhite {
		kingBB = b.WhitePieces.King
	}

	kingSq := bits.TrailingZeros64(kingBB)
	return b.IsSquareAttacked(kingSq, !isWhite)
}

func (b *Board) CanCastleKing(home int, kingSide bool, isWhite bool) bool {

	if kingSide {
		kingSideBetween := (uint64(1) << (home + 1)) | (uint64(1) << (home + 2))
		return !b.IsSquareAttacked(home+1, !isWhite) &&
			!b.IsSquareAttacked(home+2, !isWhite) &&
			kingSideBetween&b.Occupancy == 0
	}

	return false
}

func (b *Board) CanCastleQueen(home int, queenSide bool, isWhite bool) bool {

	if queenSide {
		queenSideBetween := (uint64(1) << (home - 1)) | (uint64(1) << (home - 2)) | (uint64(1) << (home - 3))
		return !b.IsSquareAttacked(home-1, !isWhite) &&
			!b.IsSquareAttacked(home-2, !isWhite) &&
			queenSideBetween&b.Occupancy == 0
	}

	return false
}
