package main

import (
	"errors"
)

type Engine struct{
	board Board
}

type Board struct{
	whitePieces Pieces
	blackPieces Pieces
}

type Pieces struct{
	pawns uint64
	rooks uint64
	knights uint64
	bishops uint64
	queen uint64
	king uint64
}

// struct for pieces
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
	bitboard *uint64
	piece Piece
	isWhite bool
	mask uint64
}

func NewEngine() *Engine {
	e := &Engine{}

	e.board = Board{}

	e.board.whitePieces = Pieces{}
	e.board.blackPieces = Pieces{}


	topSidePieces := Pieces{}
	bottomSidePieces := Pieces{}

	topSidePieces.pawns = 0x00FF000000000000
	topSidePieces.rooks = 0x8100000000000000
	topSidePieces.knights = 0x4200000000000000
	topSidePieces.bishops = 0x2400000000000000

	bottomSidePieces.pawns = 0x00000000000FF00
	bottomSidePieces.rooks = 0x0000000000000081
	bottomSidePieces.knights = 0x0000000000000042
	bottomSidePieces.bishops = 0x0000000000000024

	if playAsWhite {
		e.board.whitePieces = bottomSidePieces
		e.board.blackPieces = topSidePieces

		e.board.whitePieces.king = 0X0000000000000008
		e.board.whitePieces.queen = 0X0000000000000010

		e.board.blackPieces.king = e.board.whitePieces.king << 56
		e.board.blackPieces.queen = e.board.whitePieces.queen << 56

	} else {
		e.board.whitePieces = topSidePieces
		e.board.blackPieces = bottomSidePieces

		e.board.blackPieces.king = 0X0000000000000010
		e.board.blackPieces.queen = 0X0000000000000008

		e.board.whitePieces.king = e.board.blackPieces.king << 56
		e.board.whitePieces.queen = e.board.blackPieces.queen << 56

	}

	return e
}

func (e *Engine) GetBitBoardForSquare(x, y int) (pieceInfo *PieceInfo, err error) {

	err = nil
	pieceInfo = &PieceInfo{
		nil,
		NONE,
		false,
		uint64(0),
	}

	bitboards := []*uint64 {
		&e.board.blackPieces.pawns,
		&e.board.blackPieces.rooks,
		&e.board.blackPieces.knights,
		&e.board.blackPieces.bishops,
		&e.board.blackPieces.queen,
		&e.board.blackPieces.king,

		&e.board.whitePieces.pawns,
		&e.board.whitePieces.rooks,
		&e.board.whitePieces.knights,
		&e.board.whitePieces.bishops,
		&e.board.whitePieces.queen,
		&e.board.whitePieces.king,
	}


	if pieceInfo.mask, err = SpaceToMask(x, y); err != nil {
		return
	}

	for piece, bb := range bitboards {
		// hit
		if *bb & pieceInfo.mask != 0 {
			pieceInfo.bitboard = bb
			pieceInfo.piece = Piece(piece % 6)
			pieceInfo.isWhite = piece > 5
			return
		}
	}

	return
}

func (e *Engine) GenerateLegalMoves(piece PieceInfo) (uint64, error) {
	
	return uint64(0), nil
}


func (e *Engine) GenerateMoves(piece PieceInfo, directions [][2]int) (uint64, error) {
	
	// reverse-engineer the position from the mask
	x := int(piece.mask % 8)
	y := int(piece.mask / 8)

	moves := uint64(0)
	
	var (
		occupancy uint64
		enemyOccupancy uint64
	)

	if piece.isWhite {
		occupancy = WhiteOccupancy()
		enemyOccupancy = BlackOccupancy()
	} else {
		occupancy = BlackOccupancy()
		enemyOccupancy = WhiteOccupancy()
	}

	for _, dir := range directions {
		df := dir[0]
		dr := dir[1]

		file := x + df
		rank := y + dr

		for file >= 0 && file < 8 && rank >= 0 && rank < 8 {
			if mask, err := SpaceToMask(file, rank); err != nil {
				return uint64(0), err
			} else {

				if occupancy & mask != 0 {
					break
				}

				moves |= mask

				if enemyOccupancy & mask != 0 {
					break
				}

				file += df
				rank += dr

			}
		}
	}

	return moves, nil
}

func (e *Engine) GenerateDiagonalMoves(piece PieceInfo) (uint64, error) {
	
	directions := [][2]int{
		{1, 1},   // northeast
		{-1, 1},  // northwest
		{1, -1},  // southeast
		{-1, -1}, // southwest
	}

	return e.GenerateMoves(piece, directions)
}

func (e *Engine) GenerateLateralMoves(piece PieceInfo) (uint64, error) {
	
	directions := [][2]int{
		{1, 0},   // east
		{-1, 0},  // west
		{0, 1},   // north
		{0, -1},  // south
	}

	return e.GenerateMoves(piece, directions)
}


func (e *Engine) GeneratePawnMoves(piece PieceInfo) (uint64, error) {
	

	// I need to figure out how to make captures move differently
	// also figure out how to make directions work properly
	//
	// i.e white does not always mean down or up moving pawns

	directions := [][2]int{
		{0, 1},
	}

	return e.GenerateMoves(piece, directions)
}



func SpaceToMask(x, y int) (uint64, error) {

	if x >= 8 || x < 0 || y >= 8 || y < 0 {
		err := errors.New("Coordinates for bit board out of bounds")
		return uint64(0), err
	}

	mask := uint64(1) << ((7-y)*8 + x)

	return mask, nil
}

func WhiteOccupancy() uint64 {
	return engine.board.whitePieces.pawns |
		engine.board.whitePieces.rooks|
		engine.board.whitePieces.knights|
		engine.board.whitePieces.bishops|
		engine.board.whitePieces.queen|
		engine.board.whitePieces.king
}

func BlackOccupancy() uint64 {
	return engine.board.blackPieces.pawns |
		engine.board.blackPieces.rooks|
		engine.board.blackPieces.knights|
		engine.board.blackPieces.bishops|
		engine.board.blackPieces.queen|
		engine.board.blackPieces.king
}
