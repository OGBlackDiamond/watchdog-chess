package watchdog

import (
	"math"
	"slices"

	"github.com/OGBlackDiamond/watchdog-chess/engine"
)

func ChooseMove(e *engine.Engine, depth int, whiteToMove bool) (engine.Move, bool, error) {

	bestScore := math.Inf(-1)

	var bestMove engine.Move

	found := false

	moves, err := e.GenerateLegalMovesForPosition(whiteToMove)
	if err != nil {
		return engine.Move{}, false, err
	}

	// check or stalemate
	if len(moves) == 0 {
		return engine.Move{}, false, nil
	}

	alpha := math.Inf(-1)
	beta := math.Inf(1)

	orderMoves(e, moves, whiteToMove)

	for _, move := range moves {
		child := *e
		if _, err := child.MakeMoveUnchecked(move); err != nil {
			return engine.Move{}, false, err
		}
		score, err := Negamax(&child, depth-1, -beta, -alpha, !whiteToMove)
		score *= -1 //negamaxxing

		if err != nil {
			return engine.Move{}, false, err
		}

		if !found || score > bestScore {
			bestScore = score
			bestMove = move
			found = true
		}

		if score > alpha {
			alpha = score
		}
	}

	return bestMove, found, nil
}

func Negamax(e *engine.Engine, depth int, alpha float64, beta float64, whiteToMove bool) (float64, error) {
	if depth == 0 {
		return Evaluate(e.Board, whiteToMove)
	}

	best := math.Inf(-1)

	moves, err := e.GenerateLegalMovesForPosition(whiteToMove)
	if err != nil {
		return best, err
	}

	if len(moves) == 0 {
		if e.KingIsChecked(whiteToMove) {
			return math.Inf(-1), nil // checkmate for side to move
		}
		return 0, nil // stalemate
	}

	orderMoves(e, moves, whiteToMove)

	for _, move := range moves {
		child := *e
		if _, err := child.MakeMoveUnchecked(move); err != nil {
			return 0, err
		}

		score, err := Negamax(&child, depth-1, -beta, -alpha, !whiteToMove)
		score *= -1 //negamaxxing

		if err != nil {
			return best, err
		}

		if score > best {
			best = score
		}

		if score > alpha {
			alpha = score
		}

		if alpha >= beta {
			break
		}

	}

	return best, nil
}

func orderMoves(e *engine.Engine, moves []engine.Move, whiteToMove bool) {
	slices.SortFunc(moves, func(a, b engine.Move) int {
		return scoreMove(e, b, whiteToMove) - scoreMove(e, a, whiteToMove)
	})
}

// this is NOT position scoring
func scoreMove(e *engine.Engine, move engine.Move, whiteToMove bool) int {

	score := 0

	targetPiece, occupied := e.GetBitBoardForSquare(move.ToX, move.ToY)

	attacker, attackerFound := e.GetBitBoardForSquare(move.FromX, move.FromY)

	if occupied && targetPiece.IsWhite != whiteToMove {
		score += 10_000
		score += 10 * pieceValue(targetPiece.Piece)
		
		if attackerFound {
			score -= pieceValue(attacker.Piece)
		}
	}

	if move.Promotion != engine.NONE {
		score += 9_000 + pieceValue(move.Promotion)
	}

	return score
}

