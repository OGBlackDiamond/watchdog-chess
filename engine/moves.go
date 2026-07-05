package engine

import (
	"errors"
	"math"
	"slices"
)

func (e *Engine) MakeMove(move Move) (bool, error) {
	move = normalizeMove(move)

	if CheckBounds(move.FromX, move.FromY) || CheckBounds(move.ToX, move.ToY) {
		return false, errors.New("square out of bounds")
	}

	if move.FromX == move.ToX && move.FromY == move.ToY {
		return false, errors.New("no move to make")
	}

	fromPiece, spaceIsOccupied := e.GetBitBoardForSquare(move.FromX, move.FromY)

	if !spaceIsOccupied {
		return false, errors.New("MakeMove() failed with error: piece does not exist on from square")
	}

	occ := OccupancyInfo{
		White: e.WhiteOccupancy(),
		Black: e.BlackOccupancy(),
		All:   e.Occupancy(),
	}

	// TODO: Make this check actually mean something
	// (check for turns)
	if e.WhiteToMove != fromPiece.IsWhite {
		return false, errors.New("MakeMove() failed with error: friendly piece not selected")
	}

	toPiece, spaceIsOccupied := e.GetBitBoardForSquare(move.ToX, move.ToY)

	if !spaceIsOccupied {
		// we landed on a square with no piece
		// we need to manually generate the mask for it here
		toPiece.Mask, _ = SpaceToMask(move.ToX, move.ToY)
	}

	legalMoves := make([]Move, 0, 64)

	if err := e.GenerateLegalMovesForPiece(fromPiece, &legalMoves, occ); err != nil {
		return false, errors.New("MakeMove() failed with error: " + err.Error())
	}

	if !slices.Contains(legalMoves, move) {
		return false, errors.New("MakeMove() failed with error: illegal move")
	}

	*fromPiece.Bitboard &^= fromPiece.Mask
	*fromPiece.Bitboard |= toPiece.Mask
	if spaceIsOccupied {
		*toPiece.Bitboard &^= toPiece.Mask
	}

	if err := e.makeConditionalMove(fromPiece, move); err != nil {
		return false, err
	}
	if err := e.updateConditionalMoveState(fromPiece, move); err != nil {
		return false, err
	}

	//e.computeHash()

	return true, nil

}

func (e *Engine) makeConditionalMove(piece PieceInfo, move Move) error {

	switch piece.Piece {

	case King:

		castleDirection := float64(move.ToX - move.FromX)
		// we castled
		if math.Abs(castleDirection) == 2 {

			rookX := 0

			// castling to the right, so right side rook
			if castleDirection > 0 {
				rookX = 7
			}

			castleRookMask, ok := SpaceToMask(rookX, move.ToY)

			if !ok {
				break
			}

			rookToMask, ok := SpaceToMask(move.ToX-int(castleDirection/2), move.ToY)

			if !ok {
				break
			}

			if piece.IsWhite {
				e.Board.WhitePieces.Rooks &^= castleRookMask
				e.Board.WhitePieces.Rooks |= rookToMask
			} else {
				e.Board.BlackPieces.Rooks &^= castleRookMask
				e.Board.BlackPieces.Rooks |= rookToMask
			}
		}

	case Pawn:

		if move.Promotion != NONE {
			if !validPromotionPiece(move.Promotion) {
				return errors.New("invalid promotion piece")
			}

			toMask, ok := SpaceToMask(move.ToX, move.ToY)
			if !ok {
				break
			}

			promotionBitboard := e.pieceBitboard(move.Promotion, piece.IsWhite)
			if promotionBitboard == nil {
				return errors.New("invalid promotion piece")
			}

			*piece.Bitboard &^= toMask
			*promotionBitboard |= toMask
		}

		if piece.IsWhite {
			// if en passant happened
			if e.Board.WhitePieces.Pawns&e.enPassantTarget != 0 {
				e.Board.BlackPieces.Pawns &^= e.enPassantPieceMask
			}
		} else {
			// if en passant happened
			if e.Board.BlackPieces.Pawns&e.enPassantTarget != 0 {
				e.Board.WhitePieces.Pawns &^= e.enPassantPieceMask
			}
		}

	}

	return nil
}

// gets the bitboard for a piece
func (e *Engine) pieceBitboard(piece Piece, isWhite bool) *uint64 {
	pieces := &e.Board.BlackPieces
	if isWhite {
		pieces = &e.Board.WhitePieces
	}

	switch piece {
	case Pawn:
		return &pieces.Pawns
	case Rook:
		return &pieces.Rooks
	case Knight:
		return &pieces.Knights
	case Bishop:
		return &pieces.Bishops
	case Queen:
		return &pieces.Queen
	case King:
		return &pieces.King
	default:
		return nil
	}
}

func validPromotionPiece(piece Piece) bool {
	return piece == Queen || piece == Rook || piece == Bishop || piece == Knight
}

func (e *Engine) updateConditionalMoveState(piece PieceInfo, move Move) error {
	e.enPassantTarget = uint64(0)
	e.enPassantPieceMask = uint64(0)

	// turns pass
	e.WhiteToMove = !e.WhiteToMove

	switch piece.Piece {

	case Pawn:
		if math.Abs(float64(move.ToY-move.FromY)) == 2 {
			targetY := (move.FromY + move.ToY) / 2
			targetMask, ok := SpaceToMask(move.FromX, targetY)
			if !ok {
				break
			}

			pieceMask, ok := SpaceToMask(move.ToX, move.ToY)
			if !ok {
				break
			}

			e.enPassantTarget = targetMask
			e.enPassantPieceMask = pieceMask
		}

		return nil

	// remove castling rights if the King moves
	case King:
		if piece.IsWhite {
			e.whiteCanCastleKingSide = false
			e.whiteCanCastleQueenSide = false
		} else {
			e.blackCanCastleKingSide = false
			e.blackCanCastleQueenSide = false
		}

	case Rook:
		if piece.IsWhite {
			if move.FromX == 0 && move.FromY == 7 {
				e.whiteCanCastleQueenSide = false
			} else if move.FromX == 7 && move.FromY == 7 {
				e.whiteCanCastleKingSide = false
			}
		} else {
			if move.FromX == 0 && move.FromY == 0 {
				e.blackCanCastleQueenSide = false
			} else if move.FromX == 7 && move.FromY == 0 {
				e.blackCanCastleKingSide = false
			}
		}

	}

	return nil
}

// MakeMoveUnchecked just mutates the Board, this is used for legal move checks in the future
func (e *Engine) MakeMoveUnchecked(move Move) (bool, error) {
	move = normalizeMove(move)

	fromPiece, spaceOccupied := e.GetBitBoardForSquare(move.FromX, move.FromY)

	if !spaceOccupied {
		return false, errors.New("MakeMoveUnchecked() failed with error: piece not at from square")
	}

	toPiece, isToOccupied := e.GetBitBoardForSquare(move.ToX, move.ToY)

	if !isToOccupied {
		toPiece.Mask, _ = SpaceToMask(move.ToX, move.ToY)
	}

	*fromPiece.Bitboard &^= fromPiece.Mask
	*fromPiece.Bitboard |= toPiece.Mask
	if isToOccupied {
		*toPiece.Bitboard &^= toPiece.Mask
	}

	if err := e.makeConditionalMove(fromPiece, move); err != nil {
		return false, err
	}
	if err := e.updateConditionalMoveState(fromPiece, move); err != nil {
		return false, err
	}

	//e.computeHash()

	return true, nil
}

func normalizeMove(move Move) Move {
	if move.Promotion == Pawn {
		move.Promotion = NONE
	}

	return move
}

func (e *Engine) GetBitBoardForSquare(x, y int) (PieceInfo, bool) {

	mask, ok := SpaceToMask(x, y)

	if !ok {
		return PieceInfo{}, false
	}

	return e.pieceAtMask(mask)
}

// is this really ugly? yeah, is it fast? yeah...
func (e *Engine) pieceAtMask(mask uint64) (PieceInfo, bool) {
	if e.Board.BlackPieces.Pawns&mask != 0 {
		return PieceInfo{Bitboard: &e.Board.BlackPieces.Pawns, Piece: Pawn, IsWhite: false, Mask: mask}, true
	}
	if e.Board.BlackPieces.Rooks&mask != 0 {
		return PieceInfo{Bitboard: &e.Board.BlackPieces.Rooks, Piece: Rook, IsWhite: false, Mask: mask}, true
	}
	if e.Board.BlackPieces.Knights&mask != 0 {
		return PieceInfo{Bitboard: &e.Board.BlackPieces.Knights, Piece: Knight, IsWhite: false, Mask: mask}, true
	}
	if e.Board.BlackPieces.Bishops&mask != 0 {
		return PieceInfo{Bitboard: &e.Board.BlackPieces.Bishops, Piece: Bishop, IsWhite: false, Mask: mask}, true
	}
	if e.Board.BlackPieces.Queen&mask != 0 {
		return PieceInfo{Bitboard: &e.Board.BlackPieces.Queen, Piece: Queen, IsWhite: false, Mask: mask}, true
	}
	if e.Board.BlackPieces.King&mask != 0 {
		return PieceInfo{Bitboard: &e.Board.BlackPieces.King, Piece: King, IsWhite: false, Mask: mask}, true
	}

	if e.Board.WhitePieces.Pawns&mask != 0 {
		return PieceInfo{Bitboard: &e.Board.WhitePieces.Pawns, Piece: Pawn, IsWhite: true, Mask: mask}, true
	}
	if e.Board.WhitePieces.Rooks&mask != 0 {
		return PieceInfo{Bitboard: &e.Board.WhitePieces.Rooks, Piece: Rook, IsWhite: true, Mask: mask}, true
	}
	if e.Board.WhitePieces.Knights&mask != 0 {
		return PieceInfo{Bitboard: &e.Board.WhitePieces.Knights, Piece: Knight, IsWhite: true, Mask: mask}, true
	}
	if e.Board.WhitePieces.Bishops&mask != 0 {
		return PieceInfo{Bitboard: &e.Board.WhitePieces.Bishops, Piece: Bishop, IsWhite: true, Mask: mask}, true
	}
	if e.Board.WhitePieces.Queen&mask != 0 {
		return PieceInfo{Bitboard: &e.Board.WhitePieces.Queen, Piece: Queen, IsWhite: true, Mask: mask}, true
	}
	if e.Board.WhitePieces.King&mask != 0 {
		return PieceInfo{Bitboard: &e.Board.WhitePieces.King, Piece: King, IsWhite: true, Mask: mask}, true
	}

	return PieceInfo{}, false
}
