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

	topSidePieces := Pieces{}
	bottomSidePieces := Pieces{}

	topSidePieces.Pawns = 0x00FF000000000000
	topSidePieces.Rooks = 0x8100000000000000
	topSidePieces.Knights = 0x4200000000000000
	topSidePieces.Bishops = 0x2400000000000000

	bottomSidePieces.Pawns = 0x00000000000FF00
	bottomSidePieces.Rooks = 0x0000000000000081
	bottomSidePieces.Knights = 0x0000000000000042
	bottomSidePieces.Bishops = 0x0000000000000024

	if e.PlayAsWhite {
		e.Board.WhitePieces = bottomSidePieces
		e.Board.BlackPieces = topSidePieces

		e.Board.WhitePieces.King = 0x0000000000000010
		e.Board.WhitePieces.Queen = 0x0000000000000008

		e.Board.BlackPieces.King = e.Board.WhitePieces.King << 56
		e.Board.BlackPieces.Queen = e.Board.WhitePieces.Queen << 56

	} else {
		e.Board.WhitePieces = topSidePieces
		e.Board.BlackPieces = bottomSidePieces

		e.Board.BlackPieces.King = 0x0000000000000008
		e.Board.BlackPieces.Queen = 0x0000000000000010

		e.Board.WhitePieces.King = e.Board.BlackPieces.King << 56
		e.Board.WhitePieces.Queen = e.Board.BlackPieces.Queen << 56

	}

	e.blackCanCastleKingSide = true
	e.blackCanCastleQueenSide = true

	e.whiteCanCastleKingSide = true
	e.whiteCanCastleQueenSide = true

	e.enPassantTarget = uint64(0)
	e.enPassantPieceMask = uint64(0)

	return e
}
