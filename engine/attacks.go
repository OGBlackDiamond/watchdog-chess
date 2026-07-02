package engine

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
		for mask := uint64(1); mask != 0; mask <<= 1 {
			if pieceBoard.Board&mask == 0 {
				continue
			}

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
	if piece.Piece == King {
		return e.GenerateKingAttackMask(piece)
	}

	return e.GeneratePseudoLegalMovesForPiece(piece)
}

func (e *Engine) GenerateKingAttackMask(piece PieceInfo) (uint64, error) {
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

	return e.GenerateDirectMoves(piece, directions)
}
