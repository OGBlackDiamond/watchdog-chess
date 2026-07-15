package engine

import (
	"fmt"
	"time"

	"github.com/OGBlackDiamond/watchdog-chess/internal/board"
)

const (
	// MaxPly bounds the search depth (with headroom for future extensions /
	// quiescence). The per-ply scratch buffers are sized by this.
	MaxPly = 64

	// maxMoves bounds the number of moves in one position; no legal chess
	// position has more than 218 moves.
	maxMoves = 256

	// Infinity is greater than any achievable score.
	Infinity = 1_000_000

	// mateScore is the value of delivering checkmate. Mates found at ply p
	// score mateScore-p so the search prefers the shortest mate and, when
	// losing, the longest defense.
	mateScore = 900_000
)

// move ordering score layers (highest searched first):
// captures > promotions > killers > history (quiet moves)
const (
	captureScoreBase   = 100_000
	promotionScoreBase = 90_000
	killerScore        = 80_000
	maxHistoryScore    = killerScore - 1_000

	ttHitScore = 1_000_000
)

// Searcher owns all per-search mutable state: preallocated per-ply move and
// score buffers (so the hot path never allocates) and the killer/history move
// ordering tables. It must not be shared between goroutines - for SMP, create
// one Searcher per thread.
type Searcher struct {
	moveStack  [MaxPly][maxMoves]board.Move
	scoreStack [MaxPly][maxMoves]int

	// boardStack holds the child board copies for copy-make. A node whose
	// children live at ply p writes them into boardStack[p]. Without this,
	// escape analysis moves every `child := *b` copy to the heap because
	// negamax is recursive.
	boardStack [MaxPly]board.Board

	// killers holds, per ply, the last two quiet moves that caused a beta
	// cutoff at that ply. Quiet moves matching a killer are ordered just
	// below captures.
	killers [MaxPly][2]board.Move

	// history counts quiet beta cutoffs per side/from/to, used to order
	// quiet moves that would otherwise all tie at score 0.
	history [2][64][64]int
}

type SearchReturn struct {
	move          board.Move
	depthSearched int
	score         int
	found         bool
	err           error
}

func (e *Engine) ChooseMove(depth int, numThreads int) (board.Move, bool, error) {
	if len(e.tt.entries) == 0 {
		e.tt.Resize(64)
	}

	e.tt.generation++

	if numThreads < 1 {
		numThreads = 1
	}

	results := make(chan SearchReturn, depth)

	for range depth {
		time.Sleep(10_000)
		go func(searchDepth int) {
			results <- FindMoveAtDepth(e.b, searchDepth, &e.tt)
		}(depth)
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

		fmt.Printf("info depth %d score cp %d\n", result.depthSearched, result.score)

		if result.depthSearched > bestDepth {
			bestMove = result.move
			bestDepth = result.depthSearched
			found = true
		}
	}

	return bestMove, found, nil
}

func FindMoveAtDepth(b board.Board, depth int, tt *TT) SearchReturn {

	if depth >= MaxPly {
		depth = MaxPly - 1
	}

	// One per-depth searcher; every node below reuses its buffers.
	var s Searcher

	moves := b.GeneratePseudoLegalMoves(s.moveStack[0][:0])
	scores := s.scoreStack[0][:len(moves)]
	for i, move := range moves {
		scores[i] = s.scoreMove(&b, move, 0, board.NullMove())
	}

	bestScore := -Infinity
	bestMove := board.NullMove()
	found := false

	alpha := -Infinity
	beta := Infinity

	for i := range moves {
		move := pickNext(moves, scores, i)

		child := &s.boardStack[0]
		*child = b
		if err := child.MakeMove(move); err != nil {
			return SearchReturn{
				move:          board.NullMove(),
				depthSearched: depth,
				err:           err,
			}
		}

		// moves are pseudo-legal: discard any that leave our king in check
		if child.KingIsChecked(b.WhiteToMove) {
			continue
		}

		score := -s.negamax(child, depth-1, 1, -beta, -alpha, tt)

		if !found || score > bestScore {
			bestScore = score
			bestMove = move
			found = true
		}

		if score > alpha {
			alpha = score
		}
	}

	// found == false means no legal move exists: checkmate or stalemate
	return SearchReturn{
		move:          bestMove,
		depthSearched: depth,
		score:         bestScore,
		found:         found,
	}
}

func (s *Searcher) negamax(b *board.Board, depth int, ply int, alpha int, beta int, tt *TT) int {

	if depth <= 0 || ply >= MaxPly {
		return b.Evaluate()
	}

	origAlpha := alpha

	var ttMove board.Move

	if m, d, sco, _, bnd, hit := tt.getEntry(b.Hash); hit {
		ttMove = m
		if d >= depth {
			v := valueFromTT(sco, ply)
			switch Bound(bnd) {
			case BoundExact:
				return v
			case BoundLower:
				if v >= beta {
					return v
				}
			case BoundUpper:
				if v <= alpha {
					return v
				}
			}
		}
	}

	moves := b.GeneratePseudoLegalMoves(s.moveStack[ply][:0])
	scores := s.scoreStack[ply][:len(moves)]
	for i, move := range moves {
		scores[i] = s.scoreMove(b, move, ply, ttMove)
	}

	best := -Infinity
	bestMove := board.NullMove()
	legalMoves := 0

	for i := range moves {
		move := pickNext(moves, scores, i)

		child := &s.boardStack[ply]
		*child = *b
		if err := child.MakeMove(move); err != nil {
			continue // impossible for generated moves; skip defensively
		}

		// moves are pseudo-legal: discard any that leave our king in check
		if child.KingIsChecked(b.WhiteToMove) {
			continue
		}
		legalMoves++

		score := -s.negamax(child, depth-1, ply+1, -beta, -alpha, tt)

		if score > best {
			best = score
			bestMove = move
		}

		if score > alpha {
			alpha = score
		}

		if alpha >= beta {
			// a quiet move that refutes this position is worth trying
			// early in sibling nodes: record it as a killer and bump
			// its history score
			if isQuiet(b, move) {
				if s.killers[ply][0] != move {
					s.killers[ply][1] = s.killers[ply][0]
					s.killers[ply][0] = move
				}

				h := &s.history[sideIndex(b)][move.StartSquare()][move.TargetSquare()]
				*h += depth * depth
				if *h > maxHistoryScore {
					*h = maxHistoryScore
				}
			}
			break
		}
	}

	if legalMoves == 0 {
		if b.KingIsChecked(b.WhiteToMove) {
			return -(mateScore - ply) // checkmate: prefer shorter mates
		}
		return 0 // stalemate
	}

	var bnd Bound
	switch {
	case best <= origAlpha:
		bnd = BoundUpper
	case best >= beta:
		bnd = BoundLower
	default:
		bnd = BoundExact
	}
	tt.storeEntry(b.Hash, bestMove, depth, valueToTT(best, ply), int(bnd))

	return best
}

// pickNext performs one step of a lazy selection sort: it finds the highest
// scored move in moves[i:], swaps it (and its score) into position i, and
// returns it. Alpha-beta usually cuts off within the first few moves, so this
// beats fully sorting the move list.
func pickNext(moves []board.Move, scores []int, i int) board.Move {
	best := i
	for j := i + 1; j < len(moves); j++ {
		if scores[j] > scores[best] {
			best = j
		}
	}

	moves[i], moves[best] = moves[best], moves[i]
	scores[i], scores[best] = scores[best], scores[i]

	return moves[i]
}

// scoreMove assigns a move-ordering score. This is NOT position scoring.
func (s *Searcher) scoreMove(b *board.Board, move board.Move, ply int, ttMove board.Move) int {

	if move == ttMove {
		return ttHitScore
	}

	score := 0

	targetPiece := b.MailBox[move.TargetSquare()]

	if targetPiece.Type() != board.NONE {
		// MVV-LVA: most valuable victim, least valuable attacker
		attacker := b.MailBox[move.StartSquare()]
		score = captureScoreBase + 10*board.PieceValue(targetPiece) - board.PieceValue(attacker)
	} else if move.Flag() == board.EnPassantCaptureFlag {
		// pawn takes pawn; the target square itself is empty
		score = captureScoreBase + 9*board.PieceValue(board.Pawn)
	}

	if move.Flag() >= board.PromoteToRookFlag {
		score += promotionScoreBase + promotionValue(move.Flag())
	}

	if score != 0 {
		return score
	}

	// quiet moves: killers first, then by history
	if move == s.killers[ply][0] || move == s.killers[ply][1] {
		return killerScore
	}

	return s.history[sideIndex(b)][move.StartSquare()][move.TargetSquare()]
}

// isQuiet reports whether a move is neither a capture nor a promotion.
func isQuiet(b *board.Board, move board.Move) bool {
	return b.MailBox[move.TargetSquare()].Type() == board.NONE &&
		move.Flag() != board.EnPassantCaptureFlag &&
		move.Flag() < board.PromoteToRookFlag
}

func sideIndex(b *board.Board) int {
	if b.WhiteToMove {
		return 0
	}
	return 1
}

// promotionValue maps a promotion flag to the value of the promoted piece.
func promotionValue(flag int) int {
	switch flag {
	case board.PromoteToQueenFlag:
		return board.PieceValue(board.Queen)
	case board.PromoteToRookFlag:
		return board.PieceValue(board.Rook)
	case board.PromoteToBishopFlag:
		return board.PieceValue(board.Bishop)
	case board.PromoteToKnightFlag:
		return board.PieceValue(board.Knight)
	default:
		return 0
	}
}
