package watchdog

import (
	"math/bits"

	"github.com/OGBlackDiamond/watchdog-chess/engine"
)

const (
	pawnValue   int = 100
	knightValue int = 320
	bishopValue int = 330
	rookValue   int = 500
	queenValue  int = 900
	kingValue   int = 20_000

	startValue int = pawnValue*8 +
		knightValue*2 +
		bishopValue*2 +
		rookValue*2 +
		queenValue*2 +
		kingValue*2
)

func Evaluate(board engine.Board, whiteToMove bool) (float64, error) {

	whiteScore := evaluatePieces(board.WhitePieces, true)
	blackScore := evaluatePieces(board.BlackPieces, false)

	score := whiteScore - blackScore
	if !whiteToMove {
		score *= -1
	}

	return float64(score), nil
}

func evaluatePieces(pieces engine.Pieces, isWhite bool) int {
	score := 0

	f := 1 //material(pieces) / startValue

	score += evaluateBitBoard(pieces.Pawns, engine.Pawn, isWhite, f)
	score += evaluateBitBoard(pieces.Knights, engine.Knight, isWhite, f)
	score += evaluateBitBoard(pieces.Bishops, engine.Bishop, isWhite, f)
	score += evaluateBitBoard(pieces.Rooks, engine.Rook, isWhite, f)
	score += evaluateBitBoard(pieces.Queen, engine.Queen, isWhite, f)
	score += evaluateBitBoard(pieces.King, engine.King, isWhite, f)

	return score
}

func evaluateBitBoard(bitboard uint64, piece engine.Piece, isWhite bool, f int) int {
	score := 0

	for bitboard != 0 {
		square := bits.TrailingZeros64(bitboard)
		bitboard &= bitboard - 1

		score += pieceValue(piece)
		score += pieceSquareValue(piece, square, isWhite, f)
	}

	return score
}

func pieceSquareValue(piece engine.Piece, square int, isWhite bool, f int) int {
	index := pstIndex(square, isWhite)

	switch piece {
	case engine.Pawn:
		return lerp(pawnTableMiddle[index], pawnTableEndgame[index], f)
	case engine.Knight:
		return lerp(knightTableMiddle[index], knightTableEndgame[index], f)
	case engine.Bishop:
		return lerp(bishopTableMiddle[index], bishopTableEndgame[index], f)
	case engine.Rook:
		return lerp(rookTableMiddle[index], rookTableEndgame[index], f)
	case engine.Queen:
		return lerp(queenTableMiddle[index], queenTableEndgame[index], f)
	case engine.King:
		return lerp(kingTableMiddle[index], kingTableEndgame[index], f)
	default:
		return 0
	}
}

func material(pieces engine.Pieces) int {
	return bits.OnesCount64(pieces.Pawns)*pawnValue +
		bits.OnesCount64(pieces.Knights)*knightValue +
		bits.OnesCount64(pieces.Bishops)*bishopValue +
		bits.OnesCount64(pieces.Rooks)*rookValue +
		bits.OnesCount64(pieces.Queen)*queenValue +
		bits.OnesCount64(pieces.King)*kingValue // i'm gonna try this to see if it motivates mate
}

func pieceValue(piece engine.Piece) int {
	switch piece {
	case engine.Pawn:
		return pawnValue
	case engine.Knight:
		return knightValue
	case engine.Bishop:
		return bishopValue
	case engine.Rook:
		return rookValue
	case engine.Queen:
		return queenValue
	case engine.King:
		return kingValue
	default:
		return 0
	}
}
