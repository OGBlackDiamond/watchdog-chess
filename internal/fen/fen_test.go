package fen

import (
	"testing"

	"github.com/OGBlackDiamond/watchdog-chess/internal/board"
)

// TestIndexingAgreesWithBoard verifies that the FEN parser and the board
// package's AlgNotToSpace use the same a1 = 0 square-indexing convention.
// This guards against the two packages drifting to opposite orientations.
func TestIndexingAgreesWithBoard(t *testing.T) {
	// fen.go computes a square index as rank*8 + file where rank is 0 for
	// rank 1. AlgNotToSpace must produce the identical index.
	cases := map[string]struct{ rank, file int }{
		"a1": {0, 0},
		"h1": {0, 7},
		"a8": {7, 0},
		"h8": {7, 7},
		"e2": {1, 4},
		"e4": {3, 4},
	}
	for alg, rf := range cases {
		fenIndex := rf.rank*8 + rf.file
		if got := board.AlgNotToSpace(alg); got != fenIndex {
			t.Errorf("indexing mismatch for %s: AlgNotToSpace=%d, fen formula=%d",
				alg, got, fenIndex)
		}
	}
}

// TestParserPopulatesBitboards verifies the parser writes piece bitboards back
// to the board (guards against the copy-by-value bug where pieces were set on a
// throwaway struct copy).
func TestParserPopulatesBitboards(t *testing.T) {
	b, err := NewBoardFromFen(StartingPositionFEN)
	if err != nil {
		t.Fatalf("NewBoardFromFen error: %v", err)
	}

	bit := func(alg string) uint64 { return uint64(1) << board.AlgNotToSpace(alg) }

	white := func(p board.Piece) board.Piece { return p + board.Piece(8) }

	// white pawns on rank 2, black pawns on rank 7
	if b.Bitboards[white(board.Pawn)]&bit("e2") == 0 {
		t.Errorf("white pawn bit on e2 not set")
	}
	if b.Bitboards[board.Pawn]&bit("e7") == 0 {
		t.Errorf("black pawn bit on e7 not set")
	}
	// rooks on the corners
	if b.Bitboards[white(board.Rook)]&bit("a1") == 0 || b.Bitboards[white(board.Rook)]&bit("h1") == 0 {
		t.Errorf("white rooks not set on a1/h1")
	}
	if b.Bitboards[board.King]&bit("e8") == 0 {
		t.Errorf("black king bit on e8 not set")
	}
	if b.Bitboards[white(board.King)]&bit("e1") == 0 {
		t.Errorf("white king bit on e1 not set")
	}
	// full starting position has 16 pawns
	if got := bitCount(b.Bitboards[white(board.Pawn)] | b.Bitboards[board.Pawn]); got != 16 {
		t.Errorf("expected 16 pawns, got %d", got)
	}
}

// TestParserPopulatesMailBox verifies the mailbox is populated at the same
// square indices as the bitboards, with correct piece type and color.
func TestParserPopulatesMailBox(t *testing.T) {
	b, err := NewBoardFromFen(StartingPositionFEN)
	if err != nil {
		t.Fatalf("NewBoardFromFen error: %v", err)
	}

	checks := []struct {
		alg       string
		wantType  board.Piece
		wantWhite bool
	}{
		{"e1", board.King, true},
		{"e8", board.King, false},
		{"a1", board.Rook, true},
		{"h8", board.Rook, false},
		{"e2", board.Pawn, true},
		{"e7", board.Pawn, false},
		{"d1", board.Queen, true},
		{"b8", board.Knight, false},
	}
	for _, c := range checks {
		sq := board.AlgNotToSpace(c.alg)
		got := b.MailBox[sq]
		if got.Type() != c.wantType {
			t.Errorf("%s: MailBox type = %d, want %d", c.alg, got.Type(), c.wantType)
		}
		if got.IsWhite() != c.wantWhite {
			t.Errorf("%s: MailBox IsWhite = %v, want %v", c.alg, got.IsWhite(), c.wantWhite)
		}
	}

	// empty squares (rank 3-6) must be NONE
	for _, alg := range []string{"e4", "d5", "a3", "h6"} {
		if got := b.MailBox[board.AlgNotToSpace(alg)]; got != board.NONE {
			t.Errorf("%s: MailBox should be empty, got %d", alg, got)
		}
	}
}

func bitCount(v uint64) int {
	n := 0
	for v != 0 {
		v &= v - 1
		n++
	}
	return n
}
