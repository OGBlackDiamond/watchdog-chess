package main

import (
	"errors"
	"fmt"
	"math/bits"
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

		e.board.whitePieces.king = 0X0000000000000010
		e.board.whitePieces.queen = 0X0000000000000008

		e.board.blackPieces.king = e.board.whitePieces.king << 56
		e.board.blackPieces.queen = e.board.whitePieces.queen << 56

	} else {
		e.board.whitePieces = topSidePieces
		e.board.blackPieces = bottomSidePieces

		e.board.blackPieces.king = 0X0000000000000008
		e.board.blackPieces.queen = 0X0000000000000010

		e.board.whitePieces.king = e.board.blackPieces.king << 56
		e.board.whitePieces.queen = e.board.blackPieces.queen << 56

}

	return e
}


func (e *Engine) MakeMove(fromX, fromY int, toX, toY int) (bool, error) {

	if CheckBounds(fromX, fromY) || CheckBounds(toX, toY){
		return false, errors.New("Square out of bounds")
	}

	if fromX == toX && fromY == toY {
		return false, errors.New("No move to make")
	}

	fromPiece, fromErr := e.GetBitBoardForSquare(fromX, fromY)

	if fromErr != nil {
		return false, errors.New("MakeMove() failed with error: " + fromErr.Error())
	}

	var friendlyOccupancy uint64

	if fromPiece.isWhite {
		friendlyOccupancy = WhiteOccupancy()
	} else {
		friendlyOccupancy = BlackOccupancy()
	}


	// TODO: Make this check actually mean something
	// (check for turns)
	if fromPiece.mask & friendlyOccupancy == 0 {
		return false, errors.New("MakeMove() failed with error: friendly piece not selected")
	}

	toPiece, toErr := e.GetBitBoardForSquare(toX, toY)

	toIsEmpty := false

	if toErr != nil {
		if toPiece != nil {
			// we landed on a square with no piece
			// we need to manually generate the mask for it here
			toPiece.mask, _ = SpaceToMask(toX, toY)
			toIsEmpty = true
		} else {
			return false, errors.New("MakeMove() failed with error: " + toErr.Error())
		}
	}

	legalMoves, err := e.GenerateLegalMoves(*fromPiece)

	if err != nil {
		return false, errors.New("MakeMove() failed with error: " + err.Error())
	}

	if toPiece.mask & legalMoves == 0 {
		return false, errors.New("MakeMove() failed with error: illegal move")
	}

	if toPiece.mask & friendlyOccupancy == 0 {
		
		// empty space or a capture
		// we'll have to actually do the checks here but yk
		*fromPiece.bitboard &^= fromPiece.mask
		*fromPiece.bitboard |= toPiece.mask
		if !toIsEmpty {
			*toPiece.bitboard &^= toPiece.mask
		}

		return true, nil
	}

	return false, nil
}


func (e *Engine) GetBitBoardForSquare(x, y int) (*PieceInfo, error) {

	pieceInfo := &PieceInfo{}

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


	mask, err := SpaceToMask(x, y)

	if err != nil {
		return nil, err
	}

	for piece, bb := range bitboards {
		// hit
		if *bb & mask != 0 {
			pieceInfo.mask = mask
			pieceInfo.bitboard = bb
			pieceInfo.piece = Piece(piece % 6)
			pieceInfo.isWhite = piece > 5
			return pieceInfo, nil
		}
	}

	return pieceInfo, errors.New("Square is empty")
}

func (e *Engine) GenerateLegalMoves(piece PieceInfo) (uint64, error) {

	switch piece.piece {
	case Pawn:
		return e.GeneratePawnMoves(piece)

	case Rook:
		return e.GenerateLateralMoves(piece)

	case Knight:
		return e.GenerateKnightMoves(piece)

	case Bishop:
		return e.GenerateDiagonalMoves(piece)

	case Queen:
		lateralMask, latErr := e.GenerateLateralMoves(piece)
		diagonalMask, digErr := e.GenerateDiagonalMoves(piece)

		if latErr != nil || digErr != nil {
			return uint64(0), errors.New(latErr.Error() + " " + digErr.Error())
		}

		return lateralMask | diagonalMask, nil

	case King:
		return e.GenerateKingMoves(piece)

	}

	return uint64(0), nil
}


func (e *Engine) GenerateMoves(piece PieceInfo, directions [][2]int) (uint64, error) {
	
	x, y, err := MaskToSpace(piece.mask)

	if err != nil {
		fmt.Println("GenerateMoves() failed with: " + err.Error())
		return uint64(0), err
	}

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

		for !CheckBounds(file, rank) {
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

func (e *Engine) GenerateKnightMoves(piece PieceInfo) (uint64, error) {
	
	directions := [][2]int{
		{1, 2},
		{-1, 2},
		{-2, 1},
		{-2, -1},
		{1, -2},
		{-1, -2},
		{2, -1},
		{2, 1},
	}
	
	return e.GenerateDirectMoves(piece, directions)
}


func (e *Engine) GenerateKingMoves(piece PieceInfo) (uint64, error) {
	
	directions := [][2]int{
		{0, 1},
		{-1, 1},
		{-1, 0},
		{-1, -1},
		{0, -1},
		{1, -1},
		{1, 0},
		{1, 1},
	}
	
	return e.GenerateDirectMoves(piece, directions)
}


func (e *Engine) GenerateDirectMoves(piece PieceInfo, directions [][2]int) (uint64, error) {


	mask := uint64(0)

	x, y, _ := MaskToSpace(piece.mask)
	
	
	var (
		occupancy uint64
	)

	if piece.isWhite {
		occupancy = WhiteOccupancy()
	} else {
		occupancy = BlackOccupancy()
	}


	for _, dir := range directions {
		if move, err := SpaceToMask(x + dir[0], y + dir[1]); err != nil {
			continue // this is probably wrap around
		} else {
			if move & occupancy != 0 {
				continue
			}

			mask |= move
			
		}
	}

	return mask, nil

}

func (e *Engine) GeneratePawnMoves(piece PieceInfo) (uint64, error) {

	baseDirection := -1

	if (playAsWhite && !piece.isWhite) || (!playAsWhite && piece.isWhite) {
		baseDirection = 1
	}

	directions := [][2]int{
		{0, baseDirection},
	}

	captures := [][2]int{
		{1, baseDirection},
		{-1, baseDirection},
	}

	x, y, _ := MaskToSpace(piece.mask)

	canMoveTwice := playAsWhite && ((piece.isWhite && y == 6) || (!piece.isWhite && y == 1)) ||
		!playAsWhite && ((!piece.isWhite && y == 6) || (piece.isWhite && y == 1))

	if canMoveTwice {
		directions = append(directions, [2]int{0, baseDirection * 2})
	}

	// actually start checking and adding to a mask
	
	mask := uint64(0)

	
	occupancy := Occupancy()

	var enemyOccupancy uint64

	if piece.isWhite {
		enemyOccupancy = BlackOccupancy()
	} else {
		enemyOccupancy = WhiteOccupancy()
	}


	// check moves
	for _, dir := range directions {
		if move, err := SpaceToMask(x + dir[0], y + dir[1]); err != nil {
			continue;
			// in this case we continue and don't return
			// this could be saving us from a wrap-around
			//return uint64(0), err
		} else {
			if move & occupancy != 0 {
				continue
			}

			mask |= move
		}
	}

	// check captures
	for _, dir := range captures {
		if move, err := SpaceToMask(x + dir[0], y + dir[1]); err != nil {
			continue
			//return uint64(0), err
		} else {
			if move & enemyOccupancy == 0 {
				continue
			}

			mask |= move
		}
	}


	return mask, nil
}



func SpaceToMask(x, y int) (uint64, error) {
	if CheckBounds(x, y) {
		err := errors.New("Coordinates for bit board out of bounds")
		return uint64(0), err
	}

	mask := uint64(1) << ((7-y)*8 + x)

	return mask, nil
}

func MaskToSpace(mask uint64) (int, int, error) {

	if mask == 0 {
		return 0, 0, errors.New("Mask is empty")
	}

	if mask&(mask-1) != 0 {
		return 0, 0, errors.New("Mask has more than one bit set")
	}

	square := bits.TrailingZeros64(mask)

	x := square % 8
	y := 7 - (square / 8)

	return x, y, nil
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

func Occupancy() uint64 {
	return WhiteOccupancy() | BlackOccupancy()
}

func CheckBounds(x, y int) bool {
	return x > 7 || x < 0 || y > 7 || y < 0
}
