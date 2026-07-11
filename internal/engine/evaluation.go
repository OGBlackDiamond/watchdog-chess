package engine

import (
	"math/bits"

	"github.com/OGBlackDiamond/watchdog-chess/internal/board"
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

func Evaluate(b board.Board) int {

	score := 0

	for square, piece := range b.MailBox {

		if piece.Type() == board.NONE {
			continue
		}

		multiplier := -1
		if piece.IsWhite() {
			multiplier = 1
		}

		score += (pieceValue(piece) * multiplier)
		score += (pieceSquareValue(piece, square) * multiplier)

	}

	if !b.WhiteToMove {
		score *= -1
	}

	return score
}

func pieceSquareValue(piece board.Piece, square int) int {
	index := pstIndex(square, piece.IsWhite())

	f := 1

	switch piece.Type() {
	case board.Pawn:
		return lerp(pawnTableMiddle[index], pawnTableEndgame[index], f)
	case board.Knight:
		return lerp(knightTableMiddle[index], knightTableEndgame[index], f)
	case board.Bishop:
		return lerp(bishopTableMiddle[index], bishopTableEndgame[index], f)
	case board.Rook:
		return lerp(rookTableMiddle[index], rookTableEndgame[index], f)
	case board.Queen:
		return lerp(queenTableMiddle[index], queenTableEndgame[index], f)
	case board.King:
		return lerp(kingTableMiddle[index], kingTableEndgame[index], f)
	default:
		return 0
	}
}

func material(pieces board.Pieces) int {
	return bits.OnesCount64(pieces.Pawns)*pawnValue +
		bits.OnesCount64(pieces.Knights)*knightValue +
		bits.OnesCount64(pieces.Bishops)*bishopValue +
		bits.OnesCount64(pieces.Rooks)*rookValue +
		bits.OnesCount64(pieces.Queen)*queenValue +
		bits.OnesCount64(pieces.King)*kingValue // i'm gonna try this to see if it motivates mate
}

func pieceValue(piece board.Piece) int {
	switch piece.Type() {
	case board.Pawn:
		return pawnValue
	case board.Knight:
		return knightValue
	case board.Bishop:
		return bishopValue
	case board.Rook:
		return rookValue
	case board.Queen:
		return queenValue
	case board.King:
		return kingValue
	default:
		return 0
	}
}
