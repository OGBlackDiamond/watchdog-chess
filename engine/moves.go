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

	fromPiece, fromErr := e.GetBitBoardForSquare(move.FromX, move.FromY)

	if fromErr != nil {
		return false, errors.New("MakeMove() failed with error: " + fromErr.Error())
	}

	var friendlyOccupancy uint64

	if fromPiece.IsWhite {
		friendlyOccupancy = e.WhiteOccupancy()
	} else {
		friendlyOccupancy = e.BlackOccupancy()
	}

	// TODO: Make this check actually mean something
	// (check for turns)
	if fromPiece.Mask&friendlyOccupancy == 0 {
		return false, errors.New("MakeMove() failed with error: friendly piece not selected")
	}

	toPiece, toErr := e.GetBitBoardForSquare(move.ToX, move.ToY)

	toIsEmpty := false

	if toErr != nil {
		if toPiece != nil {
			// we landed on a square with no piece
			// we need to manually generate the mask for it here
			toPiece.Mask, _ = SpaceToMask(move.ToX, move.ToY)
			toIsEmpty = true
		} else {
			return false, errors.New("MakeMove() failed with error: " + toErr.Error())
		}
	}

	legalMoves := make([]Move, 0, 64)

	if err := e.GenerateLegalMovesForPiece(*fromPiece, &legalMoves); err != nil {
		return false, errors.New("MakeMove() failed with error: " + err.Error())
	}

	if !slices.Contains(legalMoves, move) {
		return false, errors.New("MakeMove() failed with error: illegal move")
	}

	if toPiece.Mask&friendlyOccupancy == 0 {

		*fromPiece.Bitboard &^= fromPiece.Mask
		*fromPiece.Bitboard |= toPiece.Mask
		if !toIsEmpty {
			*toPiece.Bitboard &^= toPiece.Mask
		}

		if err := e.makeConditionalMove(*fromPiece, move); err != nil {
			return false, err
		}
		if err := e.updateConditionalMoveState(*fromPiece, move); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
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

			castleRookMask, err := SpaceToMask(rookX, move.ToY)

			if err != nil {
				return err
			}

			rookToMask, toErr := SpaceToMask(move.ToX-int(castleDirection/2), move.ToY)

			if toErr != nil {
				return toErr
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

			toMask, err := SpaceToMask(move.ToX, move.ToY)
			if err != nil {
				return err
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

	switch piece.Piece {

	case Pawn:
		if math.Abs(float64(move.ToY-move.FromY)) == 2 {
			targetY := (move.FromY + move.ToY) / 2
			targetMask, err := SpaceToMask(move.FromX, targetY)
			if err != nil {
				return err
			}

			pieceMask, err := SpaceToMask(move.ToX, move.ToY)
			if err != nil {
				return err
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

// just mutates the Board, this is used for legal move checks in the future
func (e *Engine) MakeMoveUnchecked(move Move) (bool, error) {
	move = normalizeMove(move)

	fromPiece, err := e.GetBitBoardForSquare(move.FromX, move.FromY)

	if err != nil {
		return false, err
	}

	toPiece, err := e.GetBitBoardForSquare(move.ToX, move.ToY)

	toIsEmpty := false

	if err != nil {
		if toPiece != nil {
			toPiece.Mask, _ = SpaceToMask(move.ToX, move.ToY)
			toIsEmpty = true
		} else {
			return false, err
		}
	}

	*fromPiece.Bitboard &^= fromPiece.Mask
	*fromPiece.Bitboard |= toPiece.Mask
	if !toIsEmpty {
		*toPiece.Bitboard &^= toPiece.Mask
	}

	if err := e.makeConditionalMove(*fromPiece, move); err != nil {
		return false, err
	}
	if err := e.updateConditionalMoveState(*fromPiece, move); err != nil {
		return false, err
	}

	return true, nil
}

func normalizeMove(move Move) Move {
	if move.Promotion == Pawn {
		move.Promotion = NONE
	}

	return move
}

func (e *Engine) GetBitBoardForSquare(x, y int) (*PieceInfo, error) {

	pieceInfo := &PieceInfo{}

	bitboards := []*uint64{
		&e.Board.BlackPieces.Pawns,
		&e.Board.BlackPieces.Rooks,
		&e.Board.BlackPieces.Knights,
		&e.Board.BlackPieces.Bishops,
		&e.Board.BlackPieces.Queen,
		&e.Board.BlackPieces.King,

		&e.Board.WhitePieces.Pawns,
		&e.Board.WhitePieces.Rooks,
		&e.Board.WhitePieces.Knights,
		&e.Board.WhitePieces.Bishops,
		&e.Board.WhitePieces.Queen,
		&e.Board.WhitePieces.King,
	}

	mask, err := SpaceToMask(x, y)

	if err != nil {
		return nil, err
	}

	for piece, bb := range bitboards {
		// hit
		if *bb&mask != 0 {
			pieceInfo.Mask = mask
			pieceInfo.Bitboard = bb
			pieceInfo.Piece = Piece(piece % 6)
			pieceInfo.IsWhite = piece > 5
			return pieceInfo, nil
		}
	}

	return pieceInfo, errors.New("square is empty")
}
