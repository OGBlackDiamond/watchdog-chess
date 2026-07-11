package board

import (
	"math/bits"
	"sort"
	"testing"
)

// maskFromSquares builds a bitboard from a list of algebraic squares.
func maskFromSquares(t *testing.T, squares ...string) uint64 {
	t.Helper()
	var m uint64
	for _, s := range squares {
		m |= uint64(1) << AlgNotToSpace(s)
	}
	return m
}

// squaresOf returns the sorted algebraic names of every set bit in a mask,
// for readable failure messages.
func squaresOf(m uint64) []string {
	var out []string
	for bb := m; bb != 0; bb &= bb - 1 {
		sq := bits.TrailingZeros64(bb)
		file, rank, err := SquareToGrid(sq)
		if err != nil {
			out = append(out, "??")
			continue
		}
		out = append(out, string(rune('a'+file))+string(rune('1'+rank)))
	}
	sort.Strings(out)
	return out
}

// TestGridRoundTrip is the invariant that would have caught the coordinate
// convention bug: GridToMask and MaskToGrid must be exact inverses for every
// square, including argument order.
func TestGridRoundTrip(t *testing.T) {
	for sq := 0; sq < 64; sq++ {
		file, rank, err := MaskToGrid(uint64(1) << sq)
		if err != nil {
			t.Fatalf("MaskToGrid(1<<%d) error: %v", sq, err)
		}
		mask, ok := GridToMask(file, rank)
		if !ok {
			t.Fatalf("GridToMask(%d,%d) not ok for square %d", file, rank, sq)
		}
		if got := bits.TrailingZeros64(mask); got != sq {
			t.Errorf("round-trip square %d -> (file=%d,rank=%d) -> square %d",
				sq, file, rank, got)
		}
	}
}

// TestGridAnchors pins the (file, rank) convention to concrete squares.
func TestGridAnchors(t *testing.T) {
	cases := []struct {
		alg                string
		wantFile, wantRank int
	}{
		{"a1", 0, 0},
		{"h1", 7, 0},
		{"a8", 0, 7},
		{"h8", 7, 7},
		{"b7", 1, 6},
		{"f1", 5, 0},
		{"e4", 4, 3},
	}
	for _, c := range cases {
		sq := AlgNotToSpace(c.alg)
		file, rank, err := SquareToGrid(sq)
		if err != nil || file != c.wantFile || rank != c.wantRank {
			t.Errorf("%s (sq %d): got (file=%d,rank=%d,%v), want (file=%d,rank=%d)",
				c.alg, sq, file, rank, err, c.wantFile, c.wantRank)
		}
	}
}

// TestKnightAttacks verifies knight attack masks on corner, edge and center.
func TestKnightAttacks(t *testing.T) {
	wants := map[string]uint64{
		"b1": maskFromSquares(t, "a3", "c3", "d2"),
		"a1": maskFromSquares(t, "b3", "c2"),
		"e4": maskFromSquares(t, "d2", "f2", "c3", "g3", "c5", "g5", "d6", "f6"),
	}
	for from, want := range wants {
		sq := AlgNotToSpace(from)
		if got := knightAttacks[sq]; got != want {
			t.Errorf("knightAttacks[%s] = %v, want %v", from, squaresOf(got), squaresOf(want))
		}
	}
}

// TestKingAttacks verifies king attack masks on center and corner.
func TestKingAttacks(t *testing.T) {
	wants := map[string]uint64{
		"e1": maskFromSquares(t, "d1", "f1", "d2", "e2", "f2"),
		"a1": maskFromSquares(t, "b1", "a2", "b2"),
		"e4": maskFromSquares(t, "d3", "e3", "f3", "d4", "f4", "d5", "e5", "f5"),
	}
	for from, want := range wants {
		sq := AlgNotToSpace(from)
		if got := kingAttacks[sq]; got != want {
			t.Errorf("kingAttacks[%s] = %v, want %v", from, squaresOf(got), squaresOf(want))
		}
	}
}

// TestPawnAttacksFrom verifies pawn capture squares for both colors, directly
// asserting the previously-buggy b7 case resolves to a6/c6.
func TestPawnAttacksFrom(t *testing.T) {
	white := map[string]uint64{
		"e2": maskFromSquares(t, "d3", "f3"),
		"a2": maskFromSquares(t, "b3"),
		"h4": maskFromSquares(t, "g5"),
	}
	for from, want := range white {
		sq := AlgNotToSpace(from)
		if got := pawnAttacks[whiteIndex][sq]; got != want {
			t.Errorf("white pawnAttacks[%s] = %v, want %v", from, squaresOf(got), squaresOf(want))
		}
	}

	black := map[string]uint64{
		"b7": maskFromSquares(t, "a6", "c6"),
		"e7": maskFromSquares(t, "d6", "f6"),
		"a5": maskFromSquares(t, "b4"),
	}
	for from, want := range black {
		sq := AlgNotToSpace(from)
		if got := pawnAttacks[blackIndex][sq]; got != want {
			t.Errorf("black pawnAttacks[%s] = %v, want %v", from, squaresOf(got), squaresOf(want))
		}
	}
}

// TestPawnAttackersTo verifies which pawns can attack a given square.
func TestPawnAttackersTo(t *testing.T) {
	// White pawns that could capture onto d5 sit on c4/e4.
	if got, want := pawnAttackers[whiteIndex][AlgNotToSpace("d5")], maskFromSquares(t, "c4", "e4"); got != want {
		t.Errorf("white pawnAttackers[d5] = %v, want %v", squaresOf(got), squaresOf(want))
	}
	// Black pawns that could capture onto d5 sit on c6/e6.
	if got, want := pawnAttackers[blackIndex][AlgNotToSpace("d5")], maskFromSquares(t, "c6", "e6"); got != want {
		t.Errorf("black pawnAttackers[d5] = %v, want %v", squaresOf(got), squaresOf(want))
	}
}
