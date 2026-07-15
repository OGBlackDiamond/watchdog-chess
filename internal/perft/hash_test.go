package perft

import (
	"testing"

	"github.com/OGBlackDiamond/watchdog-chess/internal/board"
	"github.com/OGBlackDiamond/watchdog-chess/internal/fen"
)

// walkHashConsistency recursively walks the move tree to `depth`, asserting at
// every made move that the incrementally-maintained Zobrist hash (board.Hash,
// updated inside MakeMove) equals a from-scratch ComputeHash(). A mismatch means
// an incremental term (piece, side, castling, or en passant) diverges.
func walkHashConsistency(t *testing.T, b *board.Board, depth int) {
	t.Helper()
	if depth == 0 {
		return
	}

	moves := b.GeneratePseudoLegalMoves(make([]board.Move, 0, 64))
	for _, move := range moves {
		child := *b
		if err := child.MakeMove(move); err != nil {
			t.Fatalf("MakeMove(%s) error: %v", move.ToAlgNot(), err)
		}

		// verify on every made move, including ones that prove illegal
		if got, want := child.Hash, child.ComputeHash(); got != want {
			t.Fatalf("hash mismatch after %s: incremental 0x%016x, ComputeHash 0x%016x",
				move.ToAlgNot(), got, want)
		}

		if child.KingIsChecked(b.WhiteToMove) {
			continue
		}
		walkHashConsistency(t, &child, depth-1)
	}
}

func TestZobristIncrementalMatchesComputeHash(t *testing.T) {
	depth := 4
	if testing.Short() {
		depth = 3
	}

	for _, s := range suites {
		s := s
		t.Run(s.name, func(t *testing.T) {
			b, err := fen.NewBoardFromFen(s.fen)
			if err != nil {
				t.Fatalf("NewBoardFromFen(%q) error: %v", s.fen, err)
			}

			// the root hash from ComputeHash (via fen) must be self-consistent
			if got, want := b.Hash, b.ComputeHash(); got != want {
				t.Fatalf("root hash mismatch: field 0x%016x, ComputeHash 0x%016x", got, want)
			}

			walkHashConsistency(t, b, depth)
		})
	}
}
