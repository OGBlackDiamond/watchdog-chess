// Package board handles maintaining the board state, as well as move generation
package board

import (
	"errors"
	"math/bits"
)

// Board is a struct representing a board state - it should be initialized from the `fen` package
type Board struct {
	Bitboards [16]uint64

	Hash uint64 // Zobrist hash value

	WhiteOccupancy uint64
	BlackOccupancy uint64
	Occupancy      uint64

	// the value of the material + PST that will just update in real time
	MaterialPST int

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

const (
	whitePawnStartingRank  uint64 = 0x000000000000FF00
	whitePawnPromotionRank uint64 = 0xFF00000000000000

	blackPawnStartingRank  uint64 = 0x00FF000000000000
	blackPawnPromotionRank uint64 = 0x00000000000000FF
)

// starting king squares
const (
	whiteHome = 4  // e1
	blackHome = 60 // e8
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
		default:
			return errors.New("invalid flag in MakeMove")
		}

		if b.MailBox[startSquare].IsWhite() {
			p += Piece(8) // add the white bit
		}

		b.setSquare(targetSquare, p)
	} else {
		// if no promotion, just set the sqare with the same piece
		b.setSquare(targetSquare, startType)
	}

	oldEnPassantMask := b.EnPassantPieceMask

	b.UpdateConditionalMoveState(move, startType, targetType)

	b.clearSquare(startSquare, startType)
	b.MakeConditionalMoves(move, startType, oldEnPassantMask)

	b.WhiteToMove = !b.WhiteToMove
	b.Hash ^= zobrist.sideToMove

	return nil
}

func (b *Board) MakeMoveFromAlgNot(algString string) error {
	move, err := MoveFromAlgNot(algString, b)
	if err != nil {
		return errors.New("MakeMoveFromAlgNot failed with: " + err.Error())
	}

	return b.MakeMove(move)
}

func (b *Board) UpdateConditionalMoveState(move Move, startType Piece, targetType Piece) {

	oldCastlingMask := b.castlingRightsMask()

	switch startType.Type() {
	case King:
		if startType.IsWhite() {
			b.WhiteCanCastleKingSide = false
			b.WhiteCanCastleQueenSide = false
		} else {
			b.BlackCanCastleKingSide = false
			b.BlackCanCastleQueenSide = false
		}
	case Rook:
		startSquare := move.StartSquare()
		if startType.IsWhite() {
			switch startSquare {
			case 0:
				b.WhiteCanCastleQueenSide = false
			case 7:
				b.WhiteCanCastleKingSide = false
			}
		} else {
			switch startSquare {
			case 56:
				b.BlackCanCastleQueenSide = false
			case 63:
				b.BlackCanCastleKingSide = false
			}
		}
	}

	if targetType.Type() == Rook {
		switch move.TargetSquare() {
		case 0:
			b.WhiteCanCastleQueenSide = false
		case 7:
			b.WhiteCanCastleKingSide = false
		case 56:
			b.BlackCanCastleQueenSide = false
		case 63:
			b.BlackCanCastleKingSide = false
		}
	}

	newCastlingMask := b.castlingRightsMask()
	if oldCastlingMask != newCastlingMask {
		b.Hash ^= zobrist.castling[oldCastlingMask]
		b.Hash ^= zobrist.castling[newCastlingMask]
	}

	if b.EnPassantTarget != 0 {
		file, _, _ := MaskToGrid(b.EnPassantTarget)
		b.Hash ^= zobrist.enPassantFile[file]
	}

	b.EnPassantPieceMask = uint64(0)
	b.EnPassantTarget = uint64(0)
}

func (b *Board) MakeConditionalMoves(move Move, startType Piece, oldEnPassantMask uint64) {
	startSquare := move.StartSquare()
	targetSquare := move.TargetSquare()

	switch move.Flag() {
	case CastleFlag:
		rook := Rook
		if startType.IsWhite() {
			rook += Piece(8)
		}

		var rookFrom, rookTo int
		if targetSquare > startSquare {
			// king-side castle: rook moves from h-file to the square the king crossed.
			rookFrom = targetSquare + 1
			rookTo = targetSquare - 1
		} else {
			// queen-side castle: rook moves from a-file to the square the king crossed.
			rookFrom = targetSquare - 2
			rookTo = targetSquare + 1
		}

		b.clearSquare(rookFrom, rook)
		b.setSquare(rookTo, rook)

	case EnPassantCaptureFlag:
		capturedPawn := Pawn
		if !startType.IsWhite() {
			capturedPawn += Piece(8)
		}

		b.clearSquare(bits.TrailingZeros64(oldEnPassantMask), capturedPawn)

	case PawnTwoUpFlag:
		// the en-passant target sits behind the pawn (a1 = 0 convention).
		// A black pawn double-pushes downward in index, so behind it is a
		// higher index (+8); a white pawn pushes upward, so behind it is lower (-8).
		direction := 1
		if startType.IsWhite() {
			direction = -1
		}

		b.EnPassantPieceMask = setBit(b.EnPassantPieceMask, targetSquare)
		b.EnPassantTarget = setBit(b.EnPassantTarget, targetSquare+(8*direction))

		file, _, _ := MaskToGrid(b.EnPassantTarget)
		b.Hash ^= zobrist.enPassantFile[file]

	}
}

// returns a pointer to the bitboard for a given color+piece
func (b *Board) bitboard(p Piece) *uint64 {
	return &b.Bitboards[p]
}

func (b *Board) setSquare(sq int, p Piece) {
	bb := b.bitboard(p)
	*bb = setBit(*bb, sq)
	b.MailBox[sq] = p

	v := PieceValue(p) + pieceSquareValue(p, sq)

	b.Hash ^= zobrist.pieceSquare[p][sq]

	// update the cached Occupancy
	b.Occupancy = setBit(b.Occupancy, sq)
	if p.IsWhite() {
		b.WhiteOccupancy = setBit(b.WhiteOccupancy, sq)
		b.MaterialPST += v
	} else {
		b.BlackOccupancy = setBit(b.BlackOccupancy, sq)
		b.MaterialPST -= v
	}
}

func (b *Board) clearSquare(sq int, p Piece) {
	bb := b.bitboard(p)
	*bb = clearBit(*bb, sq)
	b.MailBox[sq] = NONE

	v := PieceValue(p) + pieceSquareValue(p, sq)

	b.Hash ^= zobrist.pieceSquare[p][sq]

	// update the cached occupancy
	b.Occupancy = clearBit(b.Occupancy, sq)
	if p.IsWhite() {
		b.WhiteOccupancy = clearBit(b.WhiteOccupancy, sq)
		b.MaterialPST -= v
	} else {
		b.BlackOccupancy = clearBit(b.BlackOccupancy, sq)
		b.MaterialPST += v
	}
}

func (b *Board) GenWhiteOccupancy() uint64 {
	occupancy := uint64(0)

	// white pieces occupy indices Pawn|0b1000 (9) through King|0b1000 (14)
	for i := int(Pawn) | 0b1000; i <= int(King)|0b1000; i++ {
		occupancy |= b.Bitboards[i]
	}

	return occupancy
}

func (b *Board) GenBlackOccupancy() uint64 {
	occupancy := uint64(0)

	// black pieces occupy indices Pawn (1) through King (6)
	for i := int(Pawn); i <= int(King); i++ {
		occupancy |= b.Bitboards[i]
	}

	return occupancy
}

// fileOf and rankOf for a square.
func fileOf(sq int) int { return sq & 7 }
func rankOf(sq int) int { return sq >> 3 }

func onBoard(f, r int) bool { return f >= 0 && f < 8 && r >= 0 && r < 8 }

func setBit(b uint64, sq int) uint64   { return b | (1 << uint(sq)) }
func clearBit(b uint64, sq int) uint64 { return b &^ (1 << uint(sq)) }
func testBit(b uint64, sq int) bool    { return b&(1<<uint(sq)) != 0 }
