package engine

import (
	"fmt"
	"math"
	"slices"

	"github.com/OGBlackDiamond/watchdog-chess/internal/board"
)

type SearchReturn struct {
	move          board.Move
	depthSearched int
	score         float64
	found         bool
	err           error
}

func ChooseMove(b *board.Board, depth int, numThreads int) (board.Move, bool, error) {
	if numThreads < 1 {
		numThreads = 1
	}

	results := make(chan SearchReturn, depth)

	defer close(results)

	for currentDepth := 1; currentDepth <= depth; currentDepth++ {
		go func(searchDepth int) {
			results <- FindMoveAtDepth(*b, searchDepth)
		}(currentDepth)
	}

	bestMove := board.NullMove()
	bestDepth := -1
	found := false

	for range depth {
		result := <-results
		if result.err != nil {
			return board.NullMove(), false, result.err
		}

		if !result.found {
			fmt.Printf("info depth %d string no legal move\n", result.depthSearched)
			continue
		}

		fmt.Printf("info depth %d score cp %d\n", result.depthSearched, int(result.score))

		if result.depthSearched > bestDepth {
			bestMove = result.move
			bestDepth = result.depthSearched
			found = true
		}
	}

	return bestMove, found, nil
}

func FindMoveAtDepth(b board.Board, depth int) SearchReturn {

	moves, err := b.GenerateLegalMovesForPosition()
	if err != nil {
		return SearchReturn{depthSearched: depth, err: err}
	}

	// check or stalemate
	if len(moves) == 0 {
		return SearchReturn{depthSearched: depth, found: false}
	}

	bestScore := math.Inf(-1)
	bestMove := board.NullMove()
	found := false

	alpha := math.Inf(-1)
	beta := math.Inf(1)

	orderMoves(&b, moves)

	for _, move := range moves {
		child := b
		if err := child.MakeMove(move); err != nil {
			return SearchReturn{
				move: board.NullMove(),
				err:  err,
			}
		}
		score, err := Negamax(&child, depth-1, -beta, -alpha)
		score *= -1 //negamaxxing

		if err != nil {
			return SearchReturn{
				move: board.NullMove(),
				err:  err,
			}
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

	return SearchReturn{
		move:          bestMove,
		depthSearched: depth,
		score:         bestScore,
		found:         found,
	}
}

func Negamax(b *board.Board, depth int, alpha float64, beta float64) (float64, error) {
	if depth == 0 {
		return float64(Evaluate(*b)), nil
	}

	best := math.Inf(-1)

	moves, err := b.GenerateLegalMovesForPosition()
	if err != nil {
		return best, err
	}

	if len(moves) == 0 {
		if b.KingIsChecked(b.WhiteToMove) {
			return math.Inf(-1), nil // checkmate for side to move
		}
		return 0, nil // stalemate
	}

	orderMoves(b, moves)

	for _, move := range moves {
		child := *b
		if err := child.MakeMove(move); err != nil {
			return 0, err
		}

		score, err := Negamax(&child, depth-1, -beta, -alpha)
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

func orderMoves(bo *board.Board, moves []board.Move) {
	slices.SortFunc(moves, func(a, b board.Move) int {
		return scoreMove(bo, b) - scoreMove(bo, a)
	})
}

// this is NOT position scoring
func scoreMove(b *board.Board, move board.Move) int {

	score := 0

	attacker := b.MailBox[move.StartSquare()]
	targetPiece := b.MailBox[move.TargetSquare()]

	if targetPiece.Type() != board.NONE && targetPiece.IsWhite() != b.WhiteToMove {
		score += 10_000
		score += 10 * pieceValue(targetPiece)

		if attacker.Type() != board.NONE {
			score -= pieceValue(attacker)
		}
	}

	if move.Flag() >= board.PromoteToRookFlag {
		score += 9_000 // eventually weight this based on the value of the promoting piece
	}

	return score
}
