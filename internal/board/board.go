// Package board handles maintaining the board state, as well as move generation
package board

import "errors"

// Board is a struct representing a board state - it should be initialized from the `fen` package
type Board struct {

	WhitePieces Pieces
	BlackPieces Pieces

	// piece lookup so we don't have to scan for bitboards
	mailBox [64]Piece

	//PlayAsWhite bool
	WhiteToMove bool

	WhiteCanCastleKingSide  bool
	WhiteCanCastleQueenSide bool
	BlackCanCastleKingSide  bool
	BlackCanCastleQueenSide bool

	EnPassantTarget    uint64
	EnPassantPieceMask uint64
}

type Pieces struct {
	Pawns   uint64
	Rooks   uint64
	Knights uint64
	Bishops uint64
	Queen   uint64
	King    uint64
}

// Piece is an enumeration of piece types
type Piece int

const (
	Pawn Piece = iota
	Rook
	Knight
	Bishop
	Queen
	King
	NONE
)



// MakeMove makes a move according to a proper move object
func (b *Board) MakeMove(move Move) error {

	startSquare := move.StartSquare()
	targetSquare := move.TargetSquare()

	b.setSquare(startSquare, b.WhiteToMove, b.mailBox[startSquare])

	targetType := b.mailBox[targetSquare]
	if targetType != NONE {
		b.clearSquare(targetSquare, !b.WhiteToMove, targetType)
	}

	return nil
}

func (b *Board) MakeMoveFromAlgNot(algString string) error {
	move, err := MoveFromAlgNot(algString)
	if err != nil {
		return errors.New("MakeMoveFromAlgNot failed with: " + err.Error())
	}

	return b.MakeMove(move)
}

// returns a pointer to the bitboard for a given color+piece
func (b *Board) bitboard(white bool, p Piece) *uint64 {
	pieces := &b.BlackPieces
    if white {
        pieces = &b.WhitePieces
    }
    switch p {
    case Pawn:   return &pieces.Pawns
    case Rook:   return &pieces.Rooks
    case Knight: return &pieces.Knights
    case Bishop: return &pieces.Bishops
    case Queen:  return &pieces.Queen
    case King:   return &pieces.King
    }
    return nil
}


func (b *Board) setSquare(sq int, white bool, p Piece) {
    *b.bitboard(white, p) |= uint64(1) << sq
    b.mailBox[sq] = p
}

func (b *Board) clearSquare(sq int, white bool, p Piece) {
    *b.bitboard(white, p) &^= uint64(1) << sq
    b.mailBox[sq] = NONE
}
