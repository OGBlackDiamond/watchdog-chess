// Package engine implements chess board state and move generation
package engine

import "math/rand"

type Engine struct {
	Board Board

	Hash uint64 // Zobrist hash value for the current board state of this engine

	PlayAsWhite bool
	WhiteToMove bool

	whiteCanCastleKingSide  bool
	whiteCanCastleQueenSide bool
	blackCanCastleKingSide  bool
	blackCanCastleQueenSide bool

	enPassantTarget    uint64
	enPassantPieceMask uint64
}

type Board struct {
	WhitePieces Pieces
	BlackPieces Pieces
}

type Move struct {
	FromX int
	FromY int
	ToX   int
	ToY   int

	Promotion Piece
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

type PieceInfo struct {
	Bitboard *uint64
	Piece    Piece
	IsWhite  bool
	Mask     uint64
}

type OccupancyInfo struct {
	White uint64
	Black uint64
	All   uint64
}

var zobrist zobristTables

type zobristTables struct {
	pieceSquare   [2][6][64]uint64
	sideToMove    uint64
	castling      [16]uint64
	enPassantFile [8]uint64
}

func NewEngine(playAsWhite bool) *Engine {
	e := &Engine{}
	e.PlayAsWhite = playAsWhite
	e.WhiteToMove = true

	e.Board = Board{}

	e.Board.WhitePieces = Pieces{}
	e.Board.BlackPieces = Pieces{}

	// The engine always stores the board in standard chess orientation:
	// bit 0 = a1, bit 63 = h8, white starts on rank 1, black on rank 8.
	// PlayAsWhite is only a GUI/display perspective flag.
	e.Board.WhitePieces.Pawns = 0x000000000000FF00
	e.Board.WhitePieces.Rooks = 0x0000000000000081
	e.Board.WhitePieces.Knights = 0x0000000000000042
	e.Board.WhitePieces.Bishops = 0x0000000000000024
	e.Board.WhitePieces.Queen = 0x0000000000000008
	e.Board.WhitePieces.King = 0x0000000000000010

	e.Board.BlackPieces.Pawns = 0x00FF000000000000
	e.Board.BlackPieces.Rooks = 0x8100000000000000
	e.Board.BlackPieces.Knights = 0x4200000000000000
	e.Board.BlackPieces.Bishops = 0x2400000000000000
	e.Board.BlackPieces.Queen = 0x0800000000000000
	e.Board.BlackPieces.King = 0x1000000000000000

	e.blackCanCastleKingSide = true
	e.blackCanCastleQueenSide = true

	e.whiteCanCastleKingSide = true
	e.whiteCanCastleQueenSide = true

	e.enPassantTarget = uint64(0)
	e.enPassantPieceMask = uint64(0)

	e.computeHash()

	return e
}

func init() {
	initZobrist()
}

// this initialized the psuedo-random numbers for zobrist hashing
func initZobrist() {
	rng := rand.New(rand.NewSource(1))

	for side := range zobrist.pieceSquare {
		for piece := range zobrist.pieceSquare[side] {
			for square := range zobrist.pieceSquare[side][piece] {
				zobrist.pieceSquare[side][piece][square] = randomNum(rng)
			}
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

// computes the hash for the current board state
func (e *Engine) computeHash() {

	hash := uint64(0)

	for side := range zobrist.pieceSquare {

		pieces := e.Board.BlackPieces
		if side == 0 {
			pieces = e.Board.WhitePieces
		}

		for piece := range zobrist.pieceSquare[side] {

			var board uint64

			switch Piece(piece) {

			case Pawn:
				board = pieces.Pawns
			case Rook:
				board = pieces.Rooks
			case Knight:
				board = pieces.Knights
			case Bishop:
				board = pieces.Bishops
			case Queen:
				board = pieces.Queen
			case King:
				board = pieces.King

			}

			for square := range zobrist.pieceSquare[side][piece] {

				if (uint64(1)<<square)&board != 0 {
					hash ^= zobrist.pieceSquare[side][piece][square]
				}

			}
		}
	}

	if !e.WhiteToMove {
		hash ^= zobrist.sideToMove
	}

	hash ^= zobrist.castling[e.castlingRightsMask()]

	if e.enPassantTarget != 0 {
		file, _, _ := MaskToSpace(e.enPassantTarget)
		hash ^= zobrist.enPassantFile[file]

		e.Hash = hash
	}

}

func (e *Engine) castlingRightsMask() int {
	rights := 0
	if e.whiteCanCastleKingSide {
		rights |= 1
	}
	if e.whiteCanCastleQueenSide {
		rights |= 2
	}
	if e.blackCanCastleKingSide {
		rights |= 4
	}
	if e.blackCanCastleQueenSide {
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
