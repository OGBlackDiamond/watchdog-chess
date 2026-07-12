// Package perft handles all testing using the PERFT testing framework.
//
// Perft (performance test) walks the legal move tree to a fixed depth and
// counts the number of leaf nodes. Comparing these counts against known-correct
// reference values is the standard way to verify move-generation correctness.
package perft

import (
	"github.com/OGBlackDiamond/watchdog-chess/internal/board"
)

// Perft counts the number of leaf nodes in the legal move tree of depth `depth`
// starting from position `b`. depth 0 counts the position itself as a single
// node.
//
// It uses copy-make recursion: Board is a flat value type, so copying it to
// produce a child position is cheap and avoids the need for an UndoMove.
// Move generation is pseudo-legal, so every move must be made and verified
// (king not left in check) before being counted - including at depth 1.
func Perft(b *board.Board, depth int) (uint64, error) {
	if depth == 0 {
		return 1, nil
	}

	moves := b.GeneratePseudoLegalMoves(make([]board.Move, 0, 64))

	var nodes uint64
	for _, move := range moves {
		child := *b
		if err := child.MakeMove(move); err != nil {
			return 0, err
		}

		if child.KingIsChecked(b.WhiteToMove) {
			continue // pseudo-legal move left the mover's king in check
		}

		if depth == 1 {
			nodes++
			continue
		}

		sub, err := Perft(&child, depth-1)
		if err != nil {
			return 0, err
		}
		nodes += sub
	}

	return nodes, nil
}

// DivideEntry pairs a root move with the number of leaf nodes reachable beneath
// it at the requested depth.
type DivideEntry struct {
	Move  string
	Nodes uint64
}

// Divide runs a perft to `depth` but reports the leaf-node count attributable to
// each individual root move, along with the total. This is the standard tool
// for locating a move-generation discrepancy: compare the per-move breakdown
// against a reference engine to find which subtree diverges.
func Divide(b *board.Board, depth int) ([]DivideEntry, uint64, error) {
	moves := b.GenerateLegalMovesForPosition()

	entries := make([]DivideEntry, 0, len(moves))
	var total uint64

	for _, move := range moves {
		var sub uint64 = 1

		if depth > 1 {
			child := *b
			if err := child.MakeMove(move); err != nil {
				return nil, 0, err
			}

			var err error
			sub, err = Perft(&child, depth-1)
			if err != nil {
				return nil, 0, err
			}
		}

		entries = append(entries, DivideEntry{Move: move.ToAlgNot(), Nodes: sub})
		total += sub
	}

	return entries, total, nil
}
