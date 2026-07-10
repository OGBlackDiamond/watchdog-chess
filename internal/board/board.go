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
type Piece uint8

func (p Piece) IsWhite() bool { return p&0b1000 != 0 }
func (p Piece) Type() Piece   { return p & 0b0111 }   // strip color
func (p Piece) IsEmpty() bool { return p == 0 }

const (
	Pawn Piece = 0b0001
	Rook Piece = 0b0010
	Knight Piece = 0b0011
	Bishop Piece = 0b0100
	Queen Piece = 0b0101
	King Piece = 0b0110
	NONE Piece = 0b0000
)



// MakeMove makes a move according to a proper move object
func (b *Board) MakeMove(move Move) error {

	startSquare := move.StartSquare()
	targetSquare := move.TargetSquare()

	moveFlag := move.Flag()

	// if there is a promotion
	if moveFlag > 3 {
	
		var p Piece

		switch moveFlag {
		case PromoteToRookFlag:
			p = Rook
		case PromoteToKnightFlag:
			p = Knight
		case PromoteToBishopFlag:
			p = Bishop
		case PromoteToQueenFlag:
			p = Queen
		}


		if b.mailBox[startSquare].IsWhite() {
			p += Piece(8) // add the white bit
		}

		b.setSquare(targetSquare, p)
	} else {
		// if no promotion, just set the sqare with the same piece
		b.setSquare(targetSquare, b.mailBox[startSquare])
	}

	b.clearSquare(startSquare, b.mailBox[startSquare])

	targetType := b.mailBox[targetSquare]
	if targetType != NONE {
		b.clearSquare(targetSquare, targetType)
	}

	if moveFlag == EnPassantCaptureFlag {

		

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
func (b *Board) bitboard(p Piece) *uint64 {
	pieces := &b.BlackPieces
    if p.IsWhite() {
        pieces = &b.WhitePieces
    }
    switch p.Type() {
    case Pawn:   return &pieces.Pawns
    case Rook:   return &pieces.Rooks
    case Knight: return &pieces.Knights
    case Bishop: return &pieces.Bishops
    case Queen:  return &pieces.Queen
    case King:   return &pieces.King
    }
    return nil
}


func (b *Board) setSquare(sq int, p Piece) {
    *b.bitboard(p) |= uint64(1) << sq
    b.mailBox[sq] = p
}

func (b *Board) clearSquare(sq int, p Piece) {
    *b.bitboard(p) &^= uint64(1) << sq
    b.mailBox[sq] = NONE
}
