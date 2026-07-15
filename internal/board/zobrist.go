package board

import "math/rand"

var zobrist zobristTables

type zobristTables struct {
	pieceSquare   [16][64]uint64
	sideToMove    uint64
	castling      [16]uint64
	enPassantFile [8]uint64
}

// when this module is used, initialize zobrist values
func init() {
	initZobrist()
}

// this initialized the psuedo-random numbers for zobrist hashing
func initZobrist() {
	rng := rand.New(rand.NewSource(1))

	for piece := range zobrist.pieceSquare {
		for square := range zobrist.pieceSquare[piece] {
			zobrist.pieceSquare[piece][square] = randomNum(rng)
		}
	}

	zobrist.sideToMove = randomNum(rng)

	for castling := range zobrist.castling {
		zobrist.castling[castling] = randomNum(rng)
	}

	for file := range zobrist.enPassantFile {
		zobrist.enPassantFile[file] = randomNum(rng)
	}
}

// ComputeHash computes the hash for the current board state
func (b *Board) ComputeHash() uint64 {

	hash := uint64(0)

	for square := range 64 {
		p := b.MailBox[square]
		if p.IsEmpty() {
			continue
		}
		hash ^= zobrist.pieceSquare[p][square]
	}

	if !b.WhiteToMove {
		hash ^= zobrist.sideToMove
	}

	hash ^= zobrist.castling[b.castlingRightsMask()]

	if b.EnPassantTarget != 0 {
		file, _, err := MaskToGrid(b.EnPassantTarget)
		hash ^= zobrist.enPassantFile[file]

		if err != nil {
			return uint64(0)
		}

	}

	return hash
}

func (b *Board) castlingRightsMask() uint8 {
	rights := uint8(0)
	if b.WhiteCanCastleKingSide {
		rights |= 1
	}
	if b.WhiteCanCastleQueenSide {
		rights |= 2
	}
	if b.BlackCanCastleKingSide {
		rights |= 4
	}
	if b.BlackCanCastleQueenSide {
		rights |= 8
	}
	return rights
}

func randomNum(rng *rand.Rand) uint64 {
	key := uint64(0)
	for key == 0 {
		key = rng.Uint64()
	}
	return key
}
