// Package board handles maintaining the board state, as well as move generation
package board

import (
	"errors"
	"math/bits"
)

// Board is a struct representing a board state - it should be initialized from the `fen` package
type Board struct {
	WhitePieces Pieces
	BlackPieces Pieces

	WhiteOccupancy uint64
	BlackOccupancy uint64
	Occupancy uint64

	// piece lookup so we don't have to scan for bitboards
	MailBox [64]Piece

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
func (p Piece) Type() Piece   { return p & 0b0111 } // strip color
func (p Piece) IsEmpty() bool { return p == 0 }

const (
	Pawn   Piece = 0b0001
	Rook   Piece = 0b0010
	Knight Piece = 0b0011
	Bishop Piece = 0b0100
	Queen  Piece = 0b0101
	King   Piece = 0b0110
	NONE   Piece = 0b0000
)

// MakeMove makes a move according to a proper move object
func (b *Board) MakeMove(move Move) error {

	startSquare := move.StartSquare()
	targetSquare := move.TargetSquare()

	moveFlag := move.Flag()

	startType := b.MailBox[startSquare]
	targetType := b.MailBox[targetSquare]

	// clear a captured piece from the target square before we overwrite it,
	// otherwise placing the moving piece there would be wiped out again
	if targetType != NONE {
		b.clearSquare(targetSquare, targetType)
	}

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

		if b.MailBox[startSquare].IsWhite() {
			p += Piece(8) // add the white bit
		}

		b.setSquare(targetSquare, p)
	} else {
		// if no promotion, just set the sqare with the same piece
		b.setSquare(targetSquare, startType)
	}

	b.clearSquare(startSquare, startType)

	switch moveFlag {
	case EnPassantCaptureFlag:

		p := Pawn

		if b.MailBox[startSquare].IsWhite() {
			p += Piece(8) // add the white bit
		}

		b.clearSquare(bits.TrailingZeros64((b.EnPassantPieceMask)), p)

	case PawnTwoUpFlag:

		// the en-passant target sits behind the pawn (a1 = 0 convention).
		// a black pawn double-pushes downward in index, so behind it is a
		// higher index (+8); a white pawn pushes upward, so behind it is a
		// lower index (-8).
		direction := 1
		if startType.IsWhite() {
			direction = -1
		}

		b.EnPassantPieceMask = setBit(b.EnPassantPieceMask, targetSquare)
		b.EnPassantTarget = setBit(b.EnPassantTarget, targetSquare + (8 * direction))

	}

	b.WhiteToMove = !b.WhiteToMove

	return nil
}

func (b *Board) MakeMoveFromAlgNot(algString string) error {
	move, err := MoveFromAlgNot(algString, b)
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
	}
	return nil
}

func (b *Board) setSquare(sq int, p Piece) {
	*b.bitboard(p) = setBit(*b.bitboard(p), sq)
	b.MailBox[sq] = p

	// update the cached Occupancy
	b.Occupancy = setBit(b.Occupancy, sq)
	if p.IsWhite() {
		b.WhiteOccupancy = setBit(b.WhiteOccupancy, sq)
	} else {
		b.BlackOccupancy = setBit(b.BlackOccupancy, sq)
	}
}

func (b *Board) clearSquare(sq int, p Piece) {
	*b.bitboard(p) = clearBit(*b.bitboard(p), sq)
	b.MailBox[sq] = NONE
	
	// update the cached occupancy
	b.Occupancy = clearBit(b.Occupancy, sq)
	if p.IsWhite() {
		b.WhiteOccupancy = clearBit(b.WhiteOccupancy, sq)
	} else {
		b.BlackOccupancy = clearBit(b.BlackOccupancy, sq)
	}
}

func (b *Board) GenWhiteOccupancy() uint64 {
	return b.WhitePieces.Pawns |
			b.WhitePieces.Knights |
			b.WhitePieces.Knights |
			b.WhitePieces.Bishops |
			b.WhitePieces.Queen |
			b.WhitePieces.King
}

func (b *Board) GenBlackOccupancy() uint64 {
	return b.WhitePieces.Pawns |
			b.WhitePieces.Knights |
			b.WhitePieces.Knights |
			b.WhitePieces.Bishops |
			b.WhitePieces.Queen |
			b.WhitePieces.King
}

// fileOf and rankOf for a square.
func fileOf(sq int) int { return sq & 7 }
func rankOf(sq int) int { return sq >> 3 }

func onBoard(f, r int) bool { return f >= 0 && f < 8 && r >= 0 && r < 8 }

func setBit(b uint64, sq int) uint64   { return b | (1 << uint(sq)) }
func clearBit(b uint64, sq int) uint64 { return b &^ (1 << uint(sq)) }
func testBit(b uint64, sq int) bool      { return b&(1<<uint(sq)) != 0 }

