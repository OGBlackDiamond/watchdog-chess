package engine

import "math/bits"

func (e *Engine) KingIsChecked(isWhite bool) bool {

	kingMask := e.Board.BlackPieces.King
	if isWhite {
		kingMask = e.Board.WhitePieces.King
	}

	return e.SquareIsAttackedBy(kingMask, !isWhite)
}

func (e *Engine) GenerateAttackMask(IsWhite bool) (uint64, error) {
	pieces := e.Board.BlackPieces
	if IsWhite {
		pieces = e.Board.WhitePieces
	}

	attackMask := uint64(0)

	pieceBoards := []struct {
		Piece Piece
		Board uint64
	}{
		{Pawn, pieces.Pawns},
		{Rook, pieces.Rooks},
		{Knight, pieces.Knights},
		{Bishop, pieces.Bishops},
		{Queen, pieces.Queen},
		{King, pieces.King},
	}

	for _, pieceBoard := range pieceBoards {
		for board := pieceBoard.Board; board != 0; board &= board - 1 {
			mask := board & -board
			pieceInfo := PieceInfo{
				Piece:   pieceBoard.Piece,
				IsWhite: IsWhite,
				Mask:    mask,
			}

			pieceAttackMask, err := e.GenerateAttackMaskForPiece(pieceInfo)
			if err != nil {
				return 0, err
			}

			attackMask |= pieceAttackMask
		}
	}

	return attackMask, nil
}

func (e *Engine) SquareIsAttackedBy(squareMask uint64, byWhite bool) bool {

	sq := bits.TrailingZeros64(squareMask)

	attackers := e.Board.BlackPieces
	if byWhite {
		attackers = e.Board.WhitePieces
	}

	color := blackIndex
	if byWhite {
		color = whiteIndex
	}

	if knightAttacks[sq]&attackers.Knights != 0 {
        return true
    }

    if kingAttacks[sq]&attackers.King != 0 {
        return true
    }

    if pawnAttackers[color][sq]&attackers.Pawns != 0 {
        return true
    }

    occupancy := e.Occupancy()

    if rookAttacks(sq, occupancy)&(attackers.Rooks|attackers.Queen) != 0 {
        return true
    }

    if bishopAttacks(sq, occupancy)&(attackers.Bishops|attackers.Queen) != 0 {
        return true
    }

    return false

}

func (e *Engine) GenerateAttackMaskForPiece(piece PieceInfo) (uint64, error) {
	fromSq := bits.TrailingZeros64(piece.Mask)

	switch piece.Piece {
	case Pawn:
		return e.GeneratePawnAttackMask(piece), nil
	case Rook:
		return rookAttacks(fromSq, e.Occupancy()), nil
	case Knight:
		return knightAttacks[fromSq], nil
	case Bishop:
		return bishopAttacks(fromSq, e.Occupancy()), nil
	case Queen:
		occupancy := e.Occupancy()
		return rookAttacks(fromSq, occupancy) | bishopAttacks(fromSq, occupancy), nil
	case King:
		return kingAttacks[fromSq], nil
	}

	return uint64(0), nil
}

func (e *Engine) GeneratePawnAttackMask(piece PieceInfo) uint64 {

	color := blackIndex
	if piece.IsWhite {
		color = whiteIndex
	}

	sq := bits.TrailingZeros64(piece.Mask)

	return pawnAttacks[color][sq]
}
