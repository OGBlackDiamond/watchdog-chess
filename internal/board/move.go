package board

import (
	"errors"
	"fmt"
	"math/bits"
)

// Move is a type that represents a move
// move data represented in it's bits
// format -> FFFFTTTTTTSSSSSS
// where F = flag, T = Target square, S = start square
type Move uint16

// named values for the flag spots
const (
	NoFlag               = 0b0000
	EnPassantCaptureFlag = 0b0001
	CastleFlag           = 0b0010
	PawnTwoUpFlag        = 0b0011
	PromoteToRookFlag    = 0b0100
	PromoteToKnightFlag  = 0b0101
	PromoteToBishopFlag  = 0b0110
	PromoteToQueenFlag   = 0b0111

	startSquareMask  uint16 = 0b0000000000111111
	targetSquareMask uint16 = 0b0000111111000000
	flagMask         uint16 = 0b1111000000000000
)

func (m *Move) StartSquare() int {
	return int(uint16(*m) & startSquareMask)
}

func (m *Move) TargetSquare() int {
	return int(uint16(*m)&targetSquareMask) >> 6
}

func (m *Move) Flag() int {
	return int(uint16(*m)&flagMask) >> 12
}

func (m *Move) ToAlgNot() string {

	start := m.StartSquare()
	target := m.TargetSquare()

	var algString string

	file, rank, err := SquareToGrid(start)

	if err != nil {
		return fmt.Sprint(NullMove())
	}

	algString += string(rune('a'+file)) + string(rune('1'+rank))

	file, rank, err = SquareToGrid(target)

	if err != nil {
		return fmt.Sprint(NullMove())
	}

	algString += string(rune('a'+file)) + string(rune('1'+rank))

	flag := m.Flag()

	// a flag exists, and it's a promotion
	if flag > 3 {

		switch flag {
		case PromoteToRookFlag:
			algString += "r"
		case PromoteToKnightFlag:
			algString += "n"
		case PromoteToBishopFlag:
			algString += "b"
		case PromoteToQueenFlag:
			algString += "q"
		}
	}

	return algString
}

// SquareToGrid converts a 0-63 square index into (file, rank) coordinates.
// a1 = 0 convention: file (a..h) = 0..7, rank (1..8) = 0..7, both increasing.
func SquareToGrid(square int) (file int, rank int, err error) {

	if square > 63 || square < 0 {
		return -1, -1, errors.New("SquareToGrid failed with: square out of bounds")
	}

	file = square % 8
	rank = square / 8

	return file, rank, nil
}

// MaskToGrid converts a single-bit mask into (file, rank) coordinates.
func MaskToGrid(mask uint64) (file int, rank int, err error) {
	return SquareToGrid(bits.TrailingZeros64(mask))
}

// GridToMask converts (file, rank) coordinates into a single-bit mask.
func GridToMask(file, rank int) (uint64, bool) {
	if !onBoard(file, rank) {
		return uint64(0), false
	}

	mask := uint64(1) << (rank*8 + file)

	return mask, true
}

func NewMove(startSquare int, targetSquare int, flag int) Move {
	return Move(startSquare | targetSquare<<6 | flag<<12)
}

func MoveFromAlgNot(algString string, b *Board) (Move, error) {

	moveLen := len(algString)
	if moveLen > 5 || moveLen < 4 {
		return NullMove(), errors.New("MakeMove failed with: move string is not valid")
	}

	var (
		startSquare  int
		targetSquare int
		flag         = NoFlag
	)

	if moveLen == 5 {

		promotionStr := algString[4:]

		switch promotionStr {
		case "r":
			flag = PromoteToRookFlag
		case "n":
			flag = PromoteToKnightFlag
		case "b":
			flag = PromoteToBishopFlag
		case "q":
			flag = PromoteToQueenFlag
		default:
			return NullMove(), errors.New("MoveFromAlgNot failed with: invalid promotion char")

		}

	}

	startString := algString[:2]
	targetString := algString[2:]

	startSquare = AlgNotToSpace(startString)
	targetSquare = AlgNotToSpace(targetString)

	// derive a special flag when we don't already have a promotion.
	if flag == NoFlag {
		startType := b.MailBox[startSquare]

		startFile := startSquare % 8
		targetFile := targetSquare % 8
		delta := targetSquare - startSquare

		switch {
		// castling: the king moves two files (e.g. e1g1, e1c1)
		case startType.Type() == King && absInt(targetFile-startFile) == 2:
			flag = CastleFlag

		// double pawn push: pawn advances two ranks
		case startType.Type() == Pawn && absInt(delta) == 16:
			flag = PawnTwoUpFlag

		// en passant: pawn moves diagonally onto the recorded target square,
		// which is empty (a normal diagonal move captures an occupant)
		case startType.Type() == Pawn &&
			b.MailBox[targetSquare] == NONE &&
			startFile != targetFile &&
			b.EnPassantTarget != 0 &&
			targetSquare == bits.TrailingZeros64(b.EnPassantTarget):
			flag = EnPassantCaptureFlag
		}
	}

	return NewMove(startSquare, targetSquare, flag), nil

}

func NullMove() Move {
	return Move(0)
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func AlgNotToSpace(algString string) int {

	runeSlice := []rune(algString)

	file := runeSlice[0]
	row := runeSlice[1]
	rank := int(row - '1')

	// a1 = 0 convention: a1 -> 0, h1 -> 7, a8 -> 56, h8 -> 63
	return rank*8 + int(file-'a')
}
