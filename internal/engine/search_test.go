package engine

import (
	"testing"

	"github.com/OGBlackDiamond/watchdog-chess/internal/fen"
)

func TestFindsMateInOne(t *testing.T) {
	// Scholar's mate: white to move, Qh5xf7#
	b, err := fen.NewBoardFromFen("r1bqkb1r/pppp1ppp/2n2n2/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 0 4")
	if err != nil {
		t.Fatal(err)
	}

	result := FindMoveAtDepth(*b, 3)
	if result.err != nil {
		t.Fatal(result.err)
	}
	if !result.found {
		t.Fatal("expected a move to be found")
	}
	if got := result.move.ToAlgNot(); got != "h5f7" {
		t.Errorf("expected mating move h5f7, got %s", got)
	}
	if want := mateScore - 1; result.score != want {
		t.Errorf("expected mate score %d, got %d", want, result.score)
	}
}

func TestPrefersShorterMate(t *testing.T) {
	// Back-rank position where white has mate-in-1 (Ra8#); a deeper search
	// must still report the shortest mate, not a longer one.
	b, err := fen.NewBoardFromFen("6k1/5ppp/8/8/8/8/8/R3K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}

	result := FindMoveAtDepth(*b, 5)
	if result.err != nil {
		t.Fatal(result.err)
	}
	if got := result.move.ToAlgNot(); got != "a1a8" {
		t.Errorf("expected a1a8 mate, got %s", got)
	}
	if want := mateScore - 1; result.score != want {
		t.Errorf("expected mate-in-1 score %d, got %d", want, result.score)
	}
}

func TestCheckmatedPositionReturnsNotFound(t *testing.T) {
	// black is already checkmated (back rank), black to move: no legal moves
	b, err := fen.NewBoardFromFen("R5k1/5ppp/8/8/8/8/8/4K3 b - - 0 1")
	if err != nil {
		t.Fatal(err)
	}

	result := FindMoveAtDepth(*b, 3)
	if result.err != nil {
		t.Fatal(result.err)
	}
	if result.found {
		t.Errorf("expected no legal move in a checkmated position, got %s", result.move.ToAlgNot())
	}
}

func TestStalematePositionReturnsNotFound(t *testing.T) {
	// classic stalemate: black king a8, white queen c7, white king c8-adjacent
	b, err := fen.NewBoardFromFen("k7/2Q5/2K5/8/8/8/8/8 b - - 0 1")
	if err != nil {
		t.Fatal(err)
	}

	result := FindMoveAtDepth(*b, 3)
	if result.err != nil {
		t.Fatal(result.err)
	}
	if result.found {
		t.Errorf("expected no legal move in a stalemated position, got %s", result.move.ToAlgNot())
	}
}

func TestSearchAvoidsIllegalMoves(t *testing.T) {
	// white king is pinned down by rook checks; only legal replies to the
	// check must be produced even though movegen is pseudo-legal
	b, err := fen.NewBoardFromFen("4k3/8/8/8/8/8/4r3/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}

	result := FindMoveAtDepth(*b, 3)
	if result.err != nil {
		t.Fatal(result.err)
	}
	if !result.found {
		t.Fatal("expected a legal move (king can capture or step aside)")
	}
	// legal replies: Kxe2, Kd1, Kf1 (d2/f2 are covered by the rook on e2)
	got := result.move.ToAlgNot()
	legal := map[string]bool{"e1e2": true, "e1d1": true, "e1f1": true}
	if !legal[got] {
		t.Errorf("search returned illegal/absurd move %s", got)
	}
}

// BenchmarkFindMoveAtDepth measures single-thread search speed and, with
// -benchmem, verifies the hot path does not allocate (expect ~1 alloc/op for
// the Searcher itself).
func BenchmarkFindMoveAtDepth(bench *testing.B) {
	b, err := fen.NewBoardFromFen(fen.StartingPositionFEN)
	if err != nil {
		bench.Fatal(err)
	}

	bench.ResetTimer()
	for range bench.N {
		result := FindMoveAtDepth(*b, 5)
		if result.err != nil {
			bench.Fatal(result.err)
		}
	}
}
