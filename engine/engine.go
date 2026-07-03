// Package engine implements chess board state and move generation
package engine

type Engine struct {
	Board       Board
	PlayAsWhite bool

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

func NewEngine(playAsWhite bool) *Engine {
	e := &Engine{}
	e.PlayAsWhite = playAsWhite

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

	return e
}
