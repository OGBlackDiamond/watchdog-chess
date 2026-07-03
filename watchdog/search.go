package watchdog

import (
	"math"

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

	for _, move := range moves {
		child := *e
		if _, err := child.MakeMove(move); err != nil {
			return engine.Move{}, false, err
		}
		score, err := Negamax(&child, depth - 1, math.Inf(-1), math.Inf(1), !whiteToMove)
		score *= -1 //negamaxxing

		if err != nil {
			return engine.Move{}, false, err
		}

		if !found || score > bestScore {
			bestScore = score
			bestMove = move
			found = true
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

	
	for _, move := range moves {
		child := *e
		if _, err := child.MakeMove(move); err != nil {
			return 0, err
		}

		score, err := Negamax(&child, depth - 1, -beta, -alpha, !whiteToMove)
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
