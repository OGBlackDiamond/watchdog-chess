package board

import "math/bits"

func (b *Board) GenerateLegalMovesForPosition() ([]Move, error) {

	moves := make([]Move, 0, 64)

	for square, piece := range b.MailBox {
		if piece.Type() == NONE {
			continue
		}
		if piece.IsWhite() != b.WhiteToMove {
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
		if err := b.GeneratePawnMoves(piece.IsWhite(), mask, moves); err != nil {
			return err
		}
	case Rook:
		if err := b.GenerateStraightMoves(piece.IsWhite(), mask, moves); err != nil {
			return err
		}
	case Knight:
		if err := b.GenerateKnightMoves(piece.IsWhite(), mask, moves); err != nil {
			return err
		}
	case Bishop:
		if err := b.GenerateDiagonalMoves(piece.IsWhite(), mask, moves); err != nil {
			return err
		}
	case Queen:
		if err := b.GenerateDiagonalMoves(piece.IsWhite(), mask, moves); err != nil {
			return err
		}
		if err := b.GenerateStraightMoves(piece.IsWhite(), mask, moves); err != nil {
			return err
		}
	case King:
		if err := b.GenerateKingMoves(piece.IsWhite(), mask, moves); err != nil {
			return err
		}
	}

	return nil
}

func (b *Board) GenerateDiagonalMoves(isWhite bool, mask uint64, moves *[]Move) error {

	startSq := bits.TrailingZeros64(mask)

	friendly := b.BlackOccupancy
	if isWhite {
		friendly = b.WhiteOccupancy
	}

	return b.legalMovesFromMask(isWhite, mask, bishopAttacks(startSq, b.Occupancy)&^friendly, moves)

}

func (b *Board) GenerateStraightMoves(isWhite bool, mask uint64, moves *[]Move) error {

	startSq := bits.TrailingZeros64(mask)

	friendly := b.BlackOccupancy
	if isWhite {
		friendly = b.WhiteOccupancy
	}

	return b.legalMovesFromMask(isWhite, mask, rookAttacks(startSq, b.Occupancy)&^friendly, moves)

}

func (b *Board) GenerateKnightMoves(isWhite bool, mask uint64, moves *[]Move) error {

	startSq := bits.TrailingZeros64(mask)

	friendly := b.BlackOccupancy
	if isWhite {
		friendly = b.WhiteOccupancy
	}

	return b.legalMovesFromMask(isWhite, mask, knightAttacks[startSq]&^friendly, moves)

}

func (b *Board) GenerateKingMoves(isWhite bool, mask uint64, moves *[]Move) error {

	startSq := bits.TrailingZeros64(mask)

	friendly := b.BlackOccupancy
	if isWhite {
		friendly = b.WhiteOccupancy
	}

	if err := b.legalMovesFromMask(isWhite, mask, kingAttacks[startSq]&^friendly, moves); err != nil {
		return err
	}

	if b.KingIsChecked(isWhite) {
		return nil // king can't castle if checked
	}

	home := blackHome
	canCastleKing := b.BlackCanCastleKingSide
	canCastleQueen := b.BlackCanCastleQueenSide
	if isWhite {
		home = whiteHome
		canCastleKing = b.WhiteCanCastleKingSide
		canCastleQueen = b.WhiteCanCastleQueenSide
	}

	if b.CanCastleKing(home, canCastleKing, isWhite) {
		if err := b.addMoveIfLegal(
			isWhite,
			NewMove(startSq, startSq+2, CastleFlag),
			moves,
		); err != nil {
			return err
		}
	}
	if b.CanCastleQueen(home, canCastleQueen, isWhite) {
		if err := b.addMoveIfLegal(
			isWhite,
			NewMove(startSq, startSq-2, CastleFlag),
			moves,
		); err != nil {
			return err
		}
	}

	return nil

}

func (b *Board) GeneratePawnMoves(isWhite bool, mask uint64, moves *[]Move) error {

	startSq := bits.TrailingZeros64(mask)

	enemyOccupancy := b.WhiteOccupancy
	index := blackIndex
	direction := -1
	startingRank := blackPawnStartingRank

	if isWhite {
		enemyOccupancy = b.BlackOccupancy
		index = whiteIndex
		direction = 1
		startingRank = whitePawnStartingRank
	}

	captureTargets := pawnAttacks[index][startSq] & enemyOccupancy

	for bb := captureTargets; bb != 0; bb &= bb - 1 {
		toSq := bits.TrailingZeros64(bb)

		_, err := b.addMoveForPawnIfLegal(isWhite, NewMove(startSq, toSq, NoFlag), moves)

		if err != nil {
			return err
		}
	}

	epTargets := pawnAttacks[index][startSq] & b.EnPassantTarget

	for bb := epTargets; bb != 0; bb &= bb - 1 {
		toSq := bits.TrailingZeros64(bb)

		_, err := b.addMoveForPawnIfLegal(isWhite, NewMove(startSq, toSq, EnPassantCaptureFlag), moves)

		if err != nil {
			return err
		}
	}

	upOneSquare := startSq + (8 * direction)
	upOneMask := setBit(uint64(0), upOneSquare)
	single := b.Occupancy&upOneMask == 0

	if single {

		didPromote, err := b.addMoveForPawnIfLegal(
			isWhite,
			NewMove(startSq, upOneSquare, NoFlag),
			moves,
		)
		if err != nil {
			return err
		}

		// don't keep adding moves if a promotion exists
		if didPromote {
			return nil
		}

		if mask&startingRank != 0 {
			upTwoSquare := upOneSquare + (8 * direction)
			upTwoMask := setBit(uint64(0), upTwoSquare)
			double := b.Occupancy&upTwoMask == 0

			if double {
				if err := b.addMoveIfLegal(
					isWhite,
					NewMove(startSq, upTwoSquare, PawnTwoUpFlag),
					moves,
				); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (b *Board) addMoveForPawnIfLegal(isWhite bool, move Move, moves *[]Move) (bool, error) {

	promotionRank := blackPawnPromotionRank

	if isWhite {
		promotionRank = whitePawnPromotionRank
	}

	targetMask := setBit(uint64(0), move.TargetSquare())
	didPromote := false

	if targetMask&promotionRank != 0 {
		didPromote = true
		for flag := PromoteToRookFlag; flag <= PromoteToQueenFlag; flag++ {
			if err := b.addMoveIfLegal(
				isWhite,
				NewMove(move.StartSquare(), move.TargetSquare(), flag),
				moves,
			); err != nil {
				return didPromote, err
			}
		}
	} else {
		if err := b.addMoveIfLegal(
			isWhite,
			move, // no promotion so return a blank move
			moves,
		); err != nil {
			return didPromote, err
		}
	}

	return didPromote, nil
}

func (b *Board) legalMovesFromMask(isWhite bool, pieceMask uint64, attackMask uint64, moves *[]Move) error {

	startSq := bits.TrailingZeros64(pieceMask)

	for bb := attackMask; bb != 0; bb &= bb - 1 {
		toSq := bits.TrailingZeros64(bb)

		if err := b.addMoveIfLegal(
			isWhite,
			NewMove(startSq, toSq, NoFlag),
			moves,
		); err != nil {
			return err
		}

	}

	return nil
}

func (b *Board) addMoveIfLegal(isWhite bool, move Move, moves *[]Move) error {

	child := *b

	if err := child.MakeMove(move); err != nil {
		return err
	}

	if child.KingIsChecked(isWhite) {
		return nil
	}

	*moves = append(*moves, move)

	return nil
}
