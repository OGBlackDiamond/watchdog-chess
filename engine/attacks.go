package engine

import "math/bits"

func (e *Engine) KingIsChecked(IsWhite bool) bool {

	kingMask := e.Board.WhitePieces.King
	if !IsWhite {
		kingMask = e.Board.BlackPieces.King
	}

	enemyAttackMask, err := e.GenerateAttackMask(!IsWhite)

	if err != nil {
		return false
	}

	return kingMask&enemyAttackMask != 0
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
	x, y, err := MaskToSpace(piece.Mask)
	if err != nil {
		return 0
	}

	direction := 1
	if piece.IsWhite {
		direction = -1
	}

	attacks := uint64(0)
	for _, dx := range [2]int{-1, 1} {
		toX := x + dx
		toY := y + direction
		if CheckBounds(toX, toY) {
			continue
		}

		mask, err := SpaceToMask(toX, toY)
		if err != nil {
			return 0
		}
		attacks |= mask
	}

	return attacks
}
