package engine

import (
	"errors"
	"math"
)

func (e *Engine) MakeMove(fromX, fromY int, toX, toY int) (bool, error) {

	if CheckBounds(fromX, fromY) || CheckBounds(toX, toY) {
		return false, errors.New("Square out of bounds")
	}

	if fromX == toX && fromY == toY {
		return false, errors.New("No move to make")
	}

	fromPiece, fromErr := e.GetBitBoardForSquare(fromX, fromY)

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

	toPiece, toErr := e.GetBitBoardForSquare(toX, toY)

	toIsEmpty := false

	if toErr != nil {
		if toPiece != nil {
			// we landed on a square with no piece
			// we need to manually generate the mask for it here
			toPiece.Mask, _ = SpaceToMask(toX, toY)
			toIsEmpty = true
		} else {
			return false, errors.New("MakeMove() failed with error: " + toErr.Error())
		}
	}

	legalMoves, err := e.GenerateLegalMovesForPiece(*fromPiece)

	if err != nil {
		return false, errors.New("MakeMove() failed with error: " + err.Error())
	}

	if toPiece.Mask&legalMoves == 0 {
		return false, errors.New("MakeMove() failed with error: illegal move")
	}

	if toPiece.Mask&friendlyOccupancy == 0 {

		// empty space or a capture
		// we'll have to actually do the checks here but yk
		*fromPiece.Bitboard &^= fromPiece.Mask
		*fromPiece.Bitboard |= toPiece.Mask
		if !toIsEmpty {
			*toPiece.Bitboard &^= toPiece.Mask
		}

		e.makeConditionalMove(*fromPiece, fromX, fromY, toX, toY)
		e.updateConditionalMoveState(*fromPiece, fromX, fromY)

		return true, nil
	}

	return false, nil
}

func (e *Engine) makeConditionalMove(piece PieceInfo, fromX, fromY, toX, toY int) error {

	switch piece.Piece {

	case King:

		castleDirection := float64(toX - fromX)
		// we castled
		if math.Abs(castleDirection) == 2 {

			rookX := 0

			// castling to the right, so right side rook
			if castleDirection > 0 {
				rookX = 7
			}

			castleRookMask, err := SpaceToMask(rookX, toY)

			if err != nil {
				return err
			}

			rookToMask, toErr := SpaceToMask(toX-int(castleDirection/2), toY)

			if toErr != nil {
				return toErr
			}

			// the castling rook is white
			if e.Board.WhitePieces.Rooks&castleRookMask != 0 {
				e.Board.WhitePieces.Rooks &^= castleRookMask
				e.Board.WhitePieces.Rooks |= rookToMask
			} else {
				e.Board.BlackPieces.Rooks &^= castleRookMask
				e.Board.BlackPieces.Rooks |= rookToMask
			}
		}

	case Pawn:

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

func (e *Engine) updateConditionalMoveState(piece PieceInfo, x, y int) error {

	switch piece.Piece {

	case Pawn:
		// we return instead of breaking becuase the enPassantTarget is reset at the end of this function
		// we don't want that to happen if the pawn actually moved
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
		// check if a rook is in any of the corners first
		bottom := y == 7
		top := y == 0
		left := x == 0
		right := x == 7

		if top {
			// if the piece is not friendly, it shouldn't be on top
			if piece.IsWhite == e.PlayAsWhite {
				break
			}

			if left {
				if e.PlayAsWhite {
					e.blackCanCastleQueenSide = false
				} else {
					e.whiteCanCastleKingSide = false
				}
			} else if right {
				if e.PlayAsWhite {
					e.blackCanCastleKingSide = false
				} else {
					e.whiteCanCastleQueenSide = false
				}
			} else {
				break
			}

		} else if bottom {
			// if the piece is friendly, it should be on bottom
			if piece.IsWhite != e.PlayAsWhite {
				break
			}

			if left {
				if e.PlayAsWhite {
					e.whiteCanCastleQueenSide = false
				} else {
					e.blackCanCastleKingSide = false
				}
			} else if right {
				if e.PlayAsWhite {
					e.whiteCanCastleKingSide = false
				} else {
					e.blackCanCastleQueenSide = false
				}
			} else {
				break
			}

		} else {
			// do nothing, rook isn't in a corner
			break
		}

	}

	e.enPassantTarget = uint64(0)
	e.enPassantPieceMask = uint64(0)

	return nil
}

// just mutates the Board, this is used for legal move checks in the future
func (e *Engine) makeMoveUnchecked(fromX, fromY, toX, toY int) (bool, error) {

	fromPiece, err := e.GetBitBoardForSquare(fromX, fromY)

	if err != nil {
		return false, err
	}

	toPiece, err := e.GetBitBoardForSquare(toX, toY)

	toIsEmpty := false

	if err != nil {
		if toPiece != nil {
			toPiece.Mask, _ = SpaceToMask(toX, toY)
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

	e.makeConditionalMove(*fromPiece, fromX, fromY, toX, toY)
	//e.updateConditionalMoveState(*fromPiece, fromX, fromY)

	return true, nil
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

	return pieceInfo, errors.New("Square is empty")
}
