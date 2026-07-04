package engine

import (
	"fmt"
	"sort"
)

// Perft counts the leaf nodes of the legal move tree at the given depth.
// It is the standard test for move generator correctness and speed:
// https://www.chessprogramming.org/Perft
//
// Reference node counts for well-known positions are published at
// https://www.chessprogramming.org/Perft_Results - if Perft disagrees with
// them, the move generator has a bug. Use PerftDivide to find it.
func (e *Engine) Perft(depth int, whiteToMove bool) (uint64, error) {
	if depth <= 0 {
		return 1, nil
	}

	moves, err := e.GenerateLegalMovesForPosition(whiteToMove)
	if err != nil {
		return 0, err
	}

	// bulk counting: at depth 1 the leaf count is just the number of legal moves
	if depth == 1 {
		return uint64(len(moves)), nil
	}

	var nodes uint64

	for _, move := range moves {
		child := *e

		ok, err := child.MakeMoveUnchecked(move)
		if err != nil {
			return 0, fmt.Errorf("perft: MakeMove(%s) failed: %w", e.MoveString(move), err)
		}
		if !ok {
			return 0, fmt.Errorf("perft: MakeMove(%s) rejected a move produced by the legal move generator", e.MoveString(move))
		}

		childNodes, err := child.Perft(depth-1, !whiteToMove)
		if err != nil {
			return 0, err
		}

		nodes += childNodes
	}

	return nodes, nil
}

// DivideEntry is the perft node count below a single root move.
type DivideEntry struct {
	Move  string
	Nodes uint64
}

// PerftDivide returns the perft count below every root move plus the total,
// sorted by move name. When Perft disagrees with a reference value, compare
// this output against a known-good engine at the same position (for example
// `stockfish` with `position fen ...` then `go perft N`), then follow the
// first differing move down one level and repeat until the offending
// position is shallow enough to inspect by hand.
func (e *Engine) PerftDivide(depth int, whiteToMove bool) ([]DivideEntry, uint64, error) {
	moves, err := e.GenerateLegalMovesForPosition(whiteToMove)
	if err != nil {
		return nil, 0, err
	}

	entries := make([]DivideEntry, 0, len(moves))
	var total uint64

	for _, move := range moves {
		child := *e

		ok, err := child.MakeMoveUnchecked(move)
		if err != nil {
			return nil, 0, fmt.Errorf("perft divide: MakeMove(%s) failed: %w", e.MoveString(move), err)
		}
		if !ok {
			return nil, 0, fmt.Errorf("perft divide: MakeMove(%s) rejected a move produced by the legal move generator", e.MoveString(move))
		}

		nodes, err := child.Perft(depth-1, !whiteToMove)
		if err != nil {
			return nil, 0, err
		}

		entries = append(entries, DivideEntry{Move: e.MoveString(move), Nodes: nodes})
		total += nodes
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Move < entries[j].Move })

	return entries, total, nil
}

// SquareName converts internal (x, y) coordinates to algebraic notation
// ("e4"). The engine uses fixed standard orientation regardless of display
// perspective, so playAsWhite is kept only for call-site compatibility.
func SquareName(x, y int, playAsWhite bool) string {
	file := x
	rank := 7 - y

	return string(rune('a'+file)) + string(rune('1'+rank))
}

// MoveString renders a move in coordinate notation ("e2e4").
func (e *Engine) MoveString(m Move) string {
	return SquareName(m.FromX, m.FromY, e.PlayAsWhite) + SquareName(m.ToX, m.ToY, e.PlayAsWhite)
}
