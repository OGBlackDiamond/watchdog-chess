package board

func (b *Board) GenerateLegalMovesForPosition() ([]Move, error) {

	moves := make([]Move, 0, 64)

	for square, piece := range b.MailBox {
		if piece.Type() == NONE {
			continue
		}

		mask := setBit(uint64(0), square)

		if err := b.GenerateLegalMovesForPiece(piece, mask, &moves); err != nil {
			return nil, err
		}

	}
	
	return moves, nil
}

func (b *Board) GenerateLegalMovesForPiece(piece Piece, mask uint64, moves *[]Move) error {

	switch piece.Type() {
	case Pawn:
	case Rook:
	case Knight:
	case Bishop:
	case Queen:
	case King:
	}

	return nil
}
