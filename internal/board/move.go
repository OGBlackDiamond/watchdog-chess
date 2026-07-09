package board

import "errors"

// move data represented in it's bits
// format -> FFFFTTTTTTSSSSSS
// where F = flag, T = Target square, S = start square
type Move uint16

// named values for the flag spots
const (
	NoFlag = 0b0000
	EnPassantCaptureFlag = 0b0001
	CastleFlag = 0b0010
	PawnTwoUpFlag = 0b0011
	PromoteToRookFlag = 0b0100
	PromoteToKnightFlag = 0b0101
	PromoteToBishopFlag = 0b0110
	PromoteToQueenFlag = 0b0111

	startSquareMask uint16 = 0b0000000000111111
	targetSquareMask uint16 = 0b0000111111000000
	flagMask uint16 = 0b1111000000000000
)

func (m *Move) StartSquare() int {
	return int(uint16(*m) & startSquareMask)
}

func (m *Move) TargetSquare() int {
	return int(uint16(*m) & targetSquareMask) >> 6
}

func (m *Move) Flag() int {
	return int(uint16(*m) & flagMask) >> 12
}

func (m *Move) ToAlgNot() string {

	start := m.StartSquare()
	target := m.TargetSquare()

	var algString string

	row, file, err := SquareToGrid(start)

	if err != nil {
		return string(NullMove())
	}

	algString += string(rune('a' + file) + rune('1' + row))

	row, file, err = SquareToGrid(target)
	
	if err != nil {
		return string(NullMove())
	}

	algString += string(rune('a' + file) + rune('1' + row))

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

func SquareToGrid(square int) (int, int, error) {

	if square > 63 || square < 0 {
		return -1, -1, errors.New("SquareToGrid failed with: square out of bounds")
	}

	row := 7 - (square / 8)
	file := square % 8

	return row, file, nil
}

func NewMove(startSquare int, targetSquare int, flag int) Move {
	return Move(startSquare | targetSquare << 6 | flag << 12)
}

func MoveFromAlgNot(algString string) (Move, error) {

	moveLen := len(algString)
	if moveLen > 5 || moveLen < 4 {
		return NullMove(), errors.New("MakeMove failed with: move string is not valid")
	}

	var (
		startSquare int
		targetSquare int
		flag = NoFlag
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


	return NewMove(startSquare, targetSquare, flag), nil

}

func NullMove() Move {
	return Move(0)
}

func AlgNotToSpace(algString string) int {

	runeSlice := []rune(algString)

	file := runeSlice[0]
	row := runeSlice[1]

	return int((7 - row) - '1') * 8 + int(file - 'a')
}
