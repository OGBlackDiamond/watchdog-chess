package engine

import (
	"fmt"
	"os"
	"testing"
)

// Test positions and expected node counts from the Chess Programming Wiki:
// https://www.chessprogramming.org/Perft_Results
//
// Any mismatch means the move generator (or MakeMove) has a bug. On failure
// the test logs a per-root-move breakdown (PerftDivide); compare it with a
// known-good engine (e.g. stockfish: `position fen ...` + `go perft N`) and
// follow the first differing move down a level to locate the bug.
const (
	kiwipeteFEN  = "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	position3FEN = "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"
	position4FEN = "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1"
	position5FEN = "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8"
	position6FEN = "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10"
)

// tiers control how much of the suite runs:
//
//	tierFast: always runs
//	tierFull: skipped with `go test -short`
//	tierDeep: only runs with PERFT_DEEP=1 (can take many minutes on a slow
//	          move generator)
const (
	tierFast = iota
	tierFull
	tierDeep
)

var perftCases = []struct {
	pos   string
	fen   string
	depth int
	want  uint64
	tier  int
}{
	{"startpos", StartingPositionFEN, 1, 20, tierFast},
	{"startpos", StartingPositionFEN, 2, 400, tierFast},
	{"startpos", StartingPositionFEN, 3, 8902, tierFast},
	{"startpos", StartingPositionFEN, 4, 197281, tierFull},
	{"startpos", StartingPositionFEN, 5, 4865609, tierFull},   // first en passant captures
	{"startpos", StartingPositionFEN, 6, 119060324, tierDeep}, // first discovered/double checks

	// kiwipete: heavy on castling, pins, and checks
	{"kiwipete", kiwipeteFEN, 1, 48, tierFast},
	{"kiwipete", kiwipeteFEN, 2, 2039, tierFast},
	{"kiwipete", kiwipeteFEN, 3, 97862, tierFull},
	{"kiwipete", kiwipeteFEN, 4, 4085603, tierFull},

	// position 3: en passant, pins, rook endgame
	{"position3", position3FEN, 1, 14, tierFast},
	{"position3", position3FEN, 2, 191, tierFast},
	{"position3", position3FEN, 3, 2812, tierFast},
	{"position3", position3FEN, 4, 43238, tierFull},
	{"position3", position3FEN, 5, 674624, tierFull},

	// position 4: promotions (including underpromotions) from depth 1
	{"position4", position4FEN, 1, 6, tierFast},
	{"position4", position4FEN, 2, 264, tierFast},
	{"position4", position4FEN, 3, 9467, tierFull},
	{"position4", position4FEN, 4, 422333, tierFull},

	// position 5: promotions and castling interactions
	{"position5", position5FEN, 1, 44, tierFast},
	{"position5", position5FEN, 2, 1486, tierFast},
	{"position5", position5FEN, 3, 62379, tierFull},

	// position 6: typical middlegame
	{"position6", position6FEN, 1, 46, tierFast},
	{"position6", position6FEN, 2, 2079, tierFast},
	{"position6", position6FEN, 3, 89890, tierFull},
}

func TestPerft(t *testing.T) {
	for _, tc := range perftCases {
		name := fmt.Sprintf("%s_depth%d", tc.pos, tc.depth)

		t.Run(name, func(t *testing.T) {
			if tc.tier >= tierFull && testing.Short() {
				t.Skip("skipped in -short mode")
			}
			if tc.tier >= tierDeep && os.Getenv("PERFT_DEEP") == "" {
				t.Skip("set PERFT_DEEP=1 to run deep perft cases")
			}

			t.Parallel()

			eng, whiteToMove, err := NewEngineFromFEN(tc.fen)
			if err != nil {
				t.Fatalf("NewEngineFromFEN(%q) failed: %v", tc.fen, err)
			}

			got, err := eng.Perft(tc.depth, whiteToMove)
			if err != nil {
				t.Fatalf("Perft(%d) returned error: %v", tc.depth, err)
			}

			if got != tc.want {
				if entries, _, derr := eng.PerftDivide(tc.depth, whiteToMove); derr == nil {
					for _, entry := range entries {
						t.Logf("  %s: %d", entry.Move, entry.Nodes)
					}
				}
				t.Errorf("perft(%d) = %d, want %d\nfen: %s", tc.depth, got, tc.want, tc.fen)
			}
		})
	}
}

func TestNewEngineFromFENStartpos(t *testing.T) {
	got, whiteToMove, err := NewEngineFromFEN(StartingPositionFEN)
	if err != nil {
		t.Fatalf("NewEngineFromFEN failed: %v", err)
	}

	if !whiteToMove {
		t.Error("expected white to move in the starting position")
	}

	want := NewEngine(true)
	if *got != *want {
		t.Errorf("engine from startpos FEN differs from NewEngine(true):\ngot  %+v\nwant %+v", *got, *want)
	}
}

func TestNewEngineUsesFixedBoardOrientation(t *testing.T) {
	playingWhite := NewEngine(true)
	playingBlack := NewEngine(false)

	if playingWhite.Board != playingBlack.Board {
		t.Errorf("NewEngine board should not depend on display perspective:\nplaying white: %+v\nplaying black: %+v", playingWhite.Board, playingBlack.Board)
	}

	if !playingWhite.PlayAsWhite {
		t.Error("NewEngine(true) should preserve PlayAsWhite=true for GUI perspective")
	}
	if playingBlack.PlayAsWhite {
		t.Error("NewEngine(false) should preserve PlayAsWhite=false for GUI perspective")
	}

	if playingWhite.Board.WhitePieces.King != 0x0000000000000010 {
		t.Errorf("white king = %#x, want e1", playingWhite.Board.WhitePieces.King)
	}
	if playingWhite.Board.BlackPieces.King != 0x1000000000000000 {
		t.Errorf("black king = %#x, want e8", playingWhite.Board.BlackPieces.King)
	}
}

func TestNewEngineFromFENEnPassant(t *testing.T) {
	// black just double-pushed e7e5, so white may capture en passant on e6
	fen := "rnbqkbnr/pppp1ppp/8/4p3/8/8/PPPPPPPP/RNBQKBNR w KQkq e6 0 2"

	eng, whiteToMove, err := NewEngineFromFEN(fen)
	if err != nil {
		t.Fatalf("NewEngineFromFEN failed: %v", err)
	}

	if !whiteToMove {
		t.Error("expected white to move")
	}

	wantTarget := uint64(1) << 44 // e6
	wantPiece := uint64(1) << 36  // e5

	if eng.enPassantTarget != wantTarget {
		t.Errorf("enPassantTarget = %#x, want %#x (e6)", eng.enPassantTarget, wantTarget)
	}
	if eng.enPassantPieceMask != wantPiece {
		t.Errorf("enPassantPieceMask = %#x, want %#x (e5)", eng.enPassantPieceMask, wantPiece)
	}
}

func TestNewEngineFromFENErrors(t *testing.T) {
	invalid := []string{
		"",
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR",               // missing fields
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP w KQkq - 0 1",           // 7 ranks
		"rnbqkbnr/pppppppp/9/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 01",   // bad digit
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNX w KQkq - 0 1",  // bad piece
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR x KQkq - 0 1",  // bad side
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQxq - 0 1",  // bad castling
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq e4 0 1", // bad ep rank
	}

	for _, fen := range invalid {
		if _, _, err := NewEngineFromFEN(fen); err == nil {
			t.Errorf("NewEngineFromFEN(%q) succeeded, want error", fen)
		}
	}
}

func TestSquareName(t *testing.T) {
	cases := []struct {
		x, y        int
		playAsWhite bool
		want        string
	}{
		{0, 0, true, "a8"},
		{7, 7, true, "h1"},
		{4, 7, true, "e1"}, // white king start, standard orientation
		{4, 6, true, "e2"},
		{4, 7, false, "e1"}, // notation stays standard even when the GUI is flipped
		{0, 0, false, "a8"},
	}

	for _, tc := range cases {
		if got := SquareName(tc.x, tc.y, tc.playAsWhite); got != tc.want {
			t.Errorf("SquareName(%d, %d, %v) = %q, want %q", tc.x, tc.y, tc.playAsWhite, got, tc.want)
		}
	}
}

// Benchmarks: run with
//
//	go test ./engine -bench . -benchmem -run '^$'
//
// BenchmarkPerft is the headline number for move generation speed - compare
// ns/op and allocs/op before and after any optimization.

func BenchmarkPerftStartposDepth3(b *testing.B) {
	eng, whiteToMove, err := NewEngineFromFEN(StartingPositionFEN)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		if _, err := eng.Perft(3, whiteToMove); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateLegalMovesStartpos(b *testing.B) {
	eng, whiteToMove, err := NewEngineFromFEN(StartingPositionFEN)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		if _, err := eng.GenerateLegalMovesForPosition(whiteToMove); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateLegalMovesKiwipete(b *testing.B) {
	eng, whiteToMove, err := NewEngineFromFEN(kiwipeteFEN)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		if _, err := eng.GenerateLegalMovesForPosition(whiteToMove); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMakeMoveStartpos(b *testing.B) {
	eng, _, err := NewEngineFromFEN(StartingPositionFEN)
	if err != nil {
		b.Fatal(err)
	}
	e2e4 := Move{FromX: 4, FromY: 6, ToX: 4, ToY: 4}

	b.ReportAllocs()
	for b.Loop() {
		child := *eng
		if _, err := child.MakeMoveUnchecked(e2e4); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkKingIsCheckedStartpos(b *testing.B) {
	eng, _, err := NewEngineFromFEN(StartingPositionFEN)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		eng.KingIsChecked(true)
	}
}
