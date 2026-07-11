package board

import "testing"

// TestAlgNotToSpace verifies parsing of algebraic squares to 0-63 indices.
// a1 = 0 convention: a1 is index 0, h1 is 7, a8 is 56, h8 is 63.
func TestAlgNotToSpace(t *testing.T) {
	cases := map[string]int{
		"a1": 0,
		"h1": 7,
		"a8": 56,
		"h8": 63,
		"e2": 12,
		"e4": 28,
	}
	for alg, want := range cases {
		if got := AlgNotToSpace(alg); got != want {
			t.Errorf("AlgNotToSpace(%q) = %d, want %d", alg, got, want)
		}
	}
}

// TestSquareToGrid verifies the mapping from index to (file, rank).
func TestSquareToGrid(t *testing.T) {
	if _, _, err := SquareToGrid(64); err == nil {
		t.Error("SquareToGrid(64) should return an out-of-bounds error")
	}
	if _, _, err := SquareToGrid(-1); err == nil {
		t.Error("SquareToGrid(-1) should return an out-of-bounds error")
	}

	// a1 = 0 convention, (file, rank) ordering.
	cases := []struct {
		square             int
		wantFile, wantRank int
	}{
		{0, 0, 0},  // a1
		{7, 7, 0},  // h1
		{56, 0, 7}, // a8
		{63, 7, 7}, // h8
		{49, 1, 6}, // b7
		{5, 5, 0},  // f1
	}
	for _, c := range cases {
		file, rank, err := SquareToGrid(c.square)
		if err != nil || file != c.wantFile || rank != c.wantRank {
			t.Errorf("SquareToGrid(%d) = (file=%d,rank=%d,%v), want (file=%d,rank=%d,nil)",
				c.square, file, rank, err, c.wantFile, c.wantRank)
		}
	}
}

// TestMoveFromAlgNotFlags verifies that promotion flags decode correctly.
func TestMoveFromAlgNotFlags(t *testing.T) {
	cases := map[string]int{
		"e2e4":  NoFlag,
		"a7a8r": PromoteToRookFlag,
		"a7a8n": PromoteToKnightFlag,
		"a7a8b": PromoteToBishopFlag,
		"a7a8q": PromoteToQueenFlag,
	}
	for alg, want := range cases {
		m, err := MoveFromAlgNot(alg, &Board{})
		if err != nil {
			t.Fatalf("MoveFromAlgNot(%q) error: %v", alg, err)
		}
		if got := m.Flag(); got != want {
			t.Errorf("MoveFromAlgNot(%q).Flag() = %d, want %d", alg, got, want)
		}
	}
}

// TestMoveFromAlgNotDetectsFlags exercises position-dependent flag detection.
func TestMoveFromAlgNotDetectsFlags(t *testing.T) {
	white := func(p Piece) Piece { return p + Piece(8) }

	t.Run("castle kingside", func(t *testing.T) {
		b := &Board{}
		b.setSquare(AlgNotToSpace("e1"), white(King))
		m, err := MoveFromAlgNot("e1g1", b)
		if err != nil {
			t.Fatal(err)
		}
		if m.Flag() != CastleFlag {
			t.Errorf("e1g1 flag = %d, want CastleFlag(%d)", m.Flag(), CastleFlag)
		}
	})

	t.Run("pawn two up", func(t *testing.T) {
		b := &Board{}
		b.setSquare(AlgNotToSpace("e2"), white(Pawn))
		m, err := MoveFromAlgNot("e2e4", b)
		if err != nil {
			t.Fatal(err)
		}
		if m.Flag() != PawnTwoUpFlag {
			t.Errorf("e2e4 flag = %d, want PawnTwoUpFlag(%d)", m.Flag(), PawnTwoUpFlag)
		}
	})

	t.Run("en passant capture", func(t *testing.T) {
		b := &Board{}
		b.setSquare(AlgNotToSpace("e5"), white(Pawn))
		// black just pushed d7d5; en-passant target is d6
		b.EnPassantTarget = uint64(1) << AlgNotToSpace("d6")
		m, err := MoveFromAlgNot("e5d6", b)
		if err != nil {
			t.Fatal(err)
		}
		if m.Flag() != EnPassantCaptureFlag {
			t.Errorf("e5d6 flag = %d, want EnPassantCaptureFlag(%d)", m.Flag(), EnPassantCaptureFlag)
		}
	})

	t.Run("normal diagonal capture is not en passant", func(t *testing.T) {
		b := &Board{}
		b.setSquare(AlgNotToSpace("e5"), white(Pawn))
		b.setSquare(AlgNotToSpace("d6"), Pawn) // black pawn present
		m, err := MoveFromAlgNot("e5d6", b)
		if err != nil {
			t.Fatal(err)
		}
		if m.Flag() != NoFlag {
			t.Errorf("e5d6 (real capture) flag = %d, want NoFlag", m.Flag())
		}
	})

	t.Run("quiet pawn push has no flag", func(t *testing.T) {
		b := &Board{}
		b.setSquare(AlgNotToSpace("e2"), white(Pawn))
		m, err := MoveFromAlgNot("e2e3", b)
		if err != nil {
			t.Fatal(err)
		}
		if m.Flag() != NoFlag {
			t.Errorf("e2e3 flag = %d, want NoFlag", m.Flag())
		}
	})
}

// TestMoveRoundTrip verifies MoveFromAlgNot and ToAlgNot are inverses.
func TestMoveRoundTrip(t *testing.T) {
	cases := []string{
		"e2e4", "a1a8", "h1h8", "a7a8q", "b2c3",
		"d7c8n", "e7e5", "a8h1", "h8a1",
	}
	for _, c := range cases {
		m, err := MoveFromAlgNot(c, &Board{})
		if err != nil {
			t.Fatalf("MoveFromAlgNot(%q) error: %v", c, err)
		}
		if got := m.ToAlgNot(); got != c {
			t.Errorf("round-trip %q: start=%d target=%d flag=%d -> %q",
				c, m.StartSquare(), m.TargetSquare(), m.Flag(), got)
		}
	}
}

// sq is a test helper: parse an algebraic square into an index using the
// board package's own convention (AlgNotToSpace).
func sq(t *testing.T, alg string) int {
	t.Helper()
	return AlgNotToSpace(alg)
}

// TestMakeMoveQuiet verifies a plain move relocates the piece.
func TestMakeMoveQuiet(t *testing.T) {
	b := &Board{}
	from, to := sq(t, "e2"), sq(t, "e4")
	b.setSquare(from, Pawn+Piece(8)) // white pawn

	if err := b.MakeMove(NewMove(from, to, NoFlag)); err != nil {
		t.Fatalf("MakeMove error: %v", err)
	}
	if b.MailBox[from] != NONE {
		t.Errorf("start square e2 should be empty, got %v", b.MailBox[from])
	}
	if b.MailBox[to].Type() != Pawn || !b.MailBox[to].IsWhite() {
		t.Errorf("target square e4 should hold a white pawn, got %v", b.MailBox[to])
	}
}

// TestMakeMoveCapture verifies a capture removes the enemy piece.
func TestMakeMoveCapture(t *testing.T) {
	b := &Board{}
	from, to := sq(t, "d4"), sq(t, "e5")
	b.setSquare(from, Pawn+Piece(8)) // white pawn
	b.setSquare(to, Pawn)            // black pawn to be captured

	if err := b.MakeMove(NewMove(from, to, NoFlag)); err != nil {
		t.Fatalf("MakeMove error: %v", err)
	}
	if !b.MailBox[to].IsWhite() || b.MailBox[to].Type() != Pawn {
		t.Errorf("e5 should hold the capturing white pawn, got %v", b.MailBox[to])
	}
	// the black pawn's bit must be cleared from its bitboard
	if b.BlackPieces.Pawns&(uint64(1)<<to) != 0 {
		t.Errorf("captured black pawn bit still set on e5")
	}
}

// TestMakeMovePromotion verifies a pawn promotes to the flagged piece.
func TestMakeMovePromotion(t *testing.T) {
	b := &Board{}
	from, to := sq(t, "a7"), sq(t, "a8")
	b.setSquare(from, Pawn+Piece(8)) // white pawn

	if err := b.MakeMove(NewMove(from, to, PromoteToQueenFlag)); err != nil {
		t.Fatalf("MakeMove error: %v", err)
	}
	if b.MailBox[to].Type() != Queen || !b.MailBox[to].IsWhite() {
		t.Errorf("a8 should hold a white queen, got %v", b.MailBox[to])
	}
}

// TestMakeMovePawnTwoUpEnPassantTarget verifies the en-passant target square
// is set behind the double-pushing pawn, for both colors.
func TestMakeMovePawnTwoUpEnPassantTarget(t *testing.T) {
	tests := []struct {
		name        string
		piece       Piece
		from, to    string
		wantEPTarge string // square the en-passant target should land on
	}{
		{"white e2e4", Pawn + Piece(8), "e2", "e4", "e3"},
		{"black e7e5", Pawn, "e7", "e5", "e6"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := &Board{}
			from, to := sq(t, tc.from), sq(t, tc.to)
			b.setSquare(from, tc.piece)

			if err := b.MakeMove(NewMove(from, to, PawnTwoUpFlag)); err != nil {
				t.Fatalf("MakeMove error: %v", err)
			}

			wantTarget := uint64(1) << sq(t, tc.wantEPTarge)
			if b.EnPassantTarget != wantTarget {
				t.Errorf("EnPassantTarget = %d (bit %d), want bit %d (%s)",
					b.EnPassantTarget, indexOfBit(b.EnPassantTarget),
					sq(t, tc.wantEPTarge), tc.wantEPTarge)
			}
			wantPiece := uint64(1) << to
			if b.EnPassantPieceMask != wantPiece {
				t.Errorf("EnPassantPieceMask = %d, want bit on %s", b.EnPassantPieceMask, tc.to)
			}
		})
	}
}

// TestMakeMoveEnPassantCapture verifies an en-passant capture removes the
// pawn identified by EnPassantPieceMask (not the empty target square).
func TestMakeMoveEnPassantCapture(t *testing.T) {
	// White pawn on e5 captures en passant onto d6, removing the black
	// pawn that just double-pushed to d5.
	b := &Board{}
	from := sq(t, "e5")   // capturing white pawn
	to := sq(t, "d6")     // empty landing square
	victim := sq(t, "d5") // black pawn to be removed

	b.setSquare(from, Pawn+Piece(8)) // white pawn
	b.setSquare(victim, Pawn)        // black pawn
	b.EnPassantPieceMask = uint64(1) << victim
	b.EnPassantTarget = uint64(1) << to

	if err := b.MakeMove(NewMove(from, to, EnPassantCaptureFlag)); err != nil {
		t.Fatalf("MakeMove error: %v", err)
	}

	if !b.MailBox[to].IsWhite() || b.MailBox[to].Type() != Pawn {
		t.Errorf("d6 should hold the capturing white pawn, got %v", b.MailBox[to])
	}
	if b.MailBox[victim] != NONE {
		t.Errorf("d5 (captured pawn) should be empty, got %v", b.MailBox[victim])
	}
	if b.BlackPieces.Pawns&(uint64(1)<<victim) != 0 {
		t.Errorf("black pawn bit on d5 should be cleared")
	}
	if b.MailBox[from] != NONE {
		t.Errorf("e5 (start) should be empty, got %v", b.MailBox[from])
	}
}

// indexOfBit returns the index of the lowest set bit, or -1 if none.
func indexOfBit(v uint64) int {
	if v == 0 {
		return -1
	}
	for i := 0; i < 64; i++ {
		if v&(uint64(1)<<i) != 0 {
			return i
		}
	}
	return -1
}
