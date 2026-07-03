package engine

import (
	"math/bits"
)

func (e *Engine) GenerateLegalMovesForPosition(whiteToMove bool) ([]Move, error) {
	pieces := e.Board.BlackPieces
	if whiteToMove {
		pieces = e.Board.WhitePieces
	}

	bitboards := []uint64{
		pieces.Pawns,
		pieces.Rooks,
		pieces.Knights,
		pieces.Bishops,
		pieces.Queen,
		pieces.King,
	}

	moves := make([]Move, 0, 64)

	for piece, bitboard := range bitboards {
		for bb := bitboard; bb != 0; bb &= bb - 1 {
			fromMask := bb & -bb

			if err := e.GenerateLegalMovesForPiece(PieceInfo{
				Piece:   Piece(piece),
				IsWhite: whiteToMove,
				Mask:    fromMask,
			},
				&moves,
			); err != nil {
				return nil, err
			}
		}
	}

	return moves, nil
}

func (e *Engine) GenerateLegalMovesForPiece(piece PieceInfo, moves *[]Move) error {

	switch piece.Piece {
	case Pawn:
		return e.GeneratePawnMoves(piece, moves)

	case Rook:
		return e.GenerateLateralMoves(piece, moves)

	case Knight:
		return e.GenerateKnightMoves(piece, moves)

	case Bishop:
		return e.GenerateDiagonalMoves(piece, moves)

	case Queen:
		if err := e.GenerateLateralMoves(piece, moves); err != nil {
			return err
		}

		if err := e.GenerateDiagonalMoves(piece, moves); err != nil {
			return err
		}

		return nil

	case King:
		return e.GenerateKingMoves(piece, moves)
	}

	return nil
}

func (e *Engine) GenerateDiagonalMoves(piece PieceInfo, moves *[]Move) error {
	fromSq := bits.TrailingZeros64(piece.Mask)
	friendly := e.BlackOccupancy()
	if piece.IsWhite {
		friendly = e.WhiteOccupancy()
	}

	return e.addLegalMovesFromMask(piece, bishopAttacks(fromSq, e.Occupancy())&^friendly, moves)
}

func (e *Engine) GenerateLateralMoves(piece PieceInfo, moves *[]Move) error {
	fromSq := bits.TrailingZeros64(piece.Mask)
	friendly := e.BlackOccupancy()
	if piece.IsWhite {
		friendly = e.WhiteOccupancy()
	}

	return e.addLegalMovesFromMask(piece, rookAttacks(fromSq, e.Occupancy())&^friendly, moves)
}

func (e *Engine) GenerateKnightMoves(piece PieceInfo, moves *[]Move) error {

	fromSq := bits.TrailingZeros64(piece.Mask)

	friendly := e.BlackOccupancy()
	if piece.IsWhite {
		friendly = e.WhiteOccupancy()
	}

	return e.addLegalMovesFromMask(piece, knightAttacks[fromSq]&^friendly, moves)
}

func (e *Engine) GenerateKingMoves(piece PieceInfo, moves *[]Move) error {

	fromSq := bits.TrailingZeros64(piece.Mask)

	friendly := e.BlackOccupancy()
	if piece.IsWhite {
		friendly = e.WhiteOccupancy()
	}

	if err := e.addLegalMovesFromMask(piece, kingAttacks[fromSq]&^friendly, moves); err != nil {
		return err
	}

	x, y, err := MaskToSpace(piece.Mask)
	if err != nil {
		return err
	}

	if piece.IsWhite {
		if e.whiteCanCastleQueenSide && e.CastlePathLegal(x, y, 0, piece.IsWhite) {
			if err := e.tryAppendLegalMove(piece, Move{FromX: x, FromY: y, ToX: x - 2, ToY: y, Promotion: NONE}, moves); err != nil {
				return err
			}
		}
		if e.whiteCanCastleKingSide && e.CastlePathLegal(x, y, 7, piece.IsWhite) {
			if err := e.tryAppendLegalMove(piece, Move{FromX: x, FromY: y, ToX: x + 2, ToY: y, Promotion: NONE}, moves); err != nil {
				return err
			}
		}
	} else {
		if e.blackCanCastleQueenSide && e.CastlePathLegal(x, y, 0, piece.IsWhite) {
			if err := e.tryAppendLegalMove(piece, Move{FromX: x, FromY: y, ToX: x - 2, ToY: y, Promotion: NONE}, moves); err != nil {
				return err
			}
		}
		if e.blackCanCastleKingSide && e.CastlePathLegal(x, y, 7, piece.IsWhite) {
			if err := e.tryAppendLegalMove(piece, Move{FromX: x, FromY: y, ToX: x + 2, ToY: y, Promotion: NONE}, moves); err != nil {
				return err
			}
		}
	}

	return nil

}

func (e *Engine) GeneratePawnMoves(piece PieceInfo, moves *[]Move) error {
	x, y, err := MaskToSpace(piece.Mask)

	if err != nil {
		return err
	}

	baseDirection := 1
	startingRank := 1

	if piece.IsWhite {
		baseDirection = -1
		startingRank = 6
	}

	occupancy := e.Occupancy()
	enemy := e.WhiteOccupancy()

	if piece.IsWhite {
		enemy = e.BlackOccupancy()
	}

	oneY := y + baseDirection

	if !CheckBounds(x, oneY) {
		oneMask, _ := SpaceToMask(x, oneY)

		if occupancy&oneMask == 0 {
			if err := e.tryAppendPawnMove(piece, x, y, x, oneY, moves); err != nil {
				return err
			}

			twoY := y + 2*baseDirection

			if y == startingRank && !CheckBounds(x, twoY) {
				twoMask, _ := SpaceToMask(x, twoY)

				if occupancy&twoMask == 0 {
					if err := e.tryAppendLegalMove(piece, Move{
						FromX: x, FromY: y,
						ToX: x, ToY: twoY, Promotion: NONE,
					}, moves); err != nil {
						return err
					}
				}
			}
		}
	}

	for _, dx := range [2]int{-1, 1} {
		toX := x + dx
		toY := y + baseDirection

		if CheckBounds(toX, toY) {
			continue
		}

		toMask, _ := SpaceToMask(toX, toY)

		normalCapture := toMask&enemy != 0
		enPassantCapture := e.enPassantTarget != 0 &&
			toMask == e.enPassantTarget &&
			e.enPassantPieceMask&enemy != 0

		if normalCapture || enPassantCapture {
			if err := e.tryAppendPawnMove(piece, x, y, toX, toY, moves); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Engine) tryAppendPawnMove(piece PieceInfo, fromX, fromY, toX, toY int, moves *[]Move) error {
	if e.isPromotionRank(toY, piece.IsWhite) {
		for _, promotion := range []Piece{Queen, Rook, Bishop, Knight} {
			if err := e.tryAppendLegalMove(piece, Move{
				FromX:     fromX,
				FromY:     fromY,
				ToX:       toX,
				ToY:       toY,
				Promotion: promotion,
			}, moves); err != nil {
				return err
			}
		}

		return nil
	}

	return e.tryAppendLegalMove(piece, Move{
		FromX:     fromX,
		FromY:     fromY,
		ToX:       toX,
		ToY:       toY,
		Promotion: NONE,
	}, moves)
}

func (e *Engine) isPromotionRank(y int, isWhite bool) bool {
	if isWhite {
		return y == 0
	}

	return y == 7
}

func (e *Engine) addLegalMovesFromMask(piece PieceInfo, mask uint64, moves *[]Move) error {
	fromX, fromY, err := MaskToSpace(piece.Mask)
	if err != nil {
		return err
	}

	for bb := mask; bb != 0; bb &= bb - 1 {
		toMask := bb & -bb

		toX, toY, err := MaskToSpace(toMask)

		if err != nil {
			return err
		}

		if err := e.tryAppendLegalMove(
			piece,
			Move{
				FromX:     fromX,
				FromY:     fromY,
				ToX:       toX,
				ToY:       toY,
				Promotion: NONE,
			},
			moves,
		); err != nil {
			return err
		}
	}

	return nil

}

func (e *Engine) tryAppendLegalMove(piece PieceInfo, move Move, moves *[]Move) error {
	child := *e

	if _, err := child.MakeMoveUnchecked(move); err != nil {
		return nil
	}

	if child.KingIsChecked(piece.IsWhite) {
		return nil
	}

	*moves = append(*moves, move)
	return nil
}
