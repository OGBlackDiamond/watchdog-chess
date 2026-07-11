package board

import "testing"

// placeStartingPosition sets up the standard chess start position directly via
// setSquare, avoiding an import cycle with the fen package.
func placeStartingPosition(t *testing.T) *Board {
	t.Helper()
	b := &Board{}
	w := func(p Piece) Piece { return p + Piece(8) }

	// White back rank (rank 1) and pawns (rank 2).
	back := []Piece{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}
	for file, p := range back {
		b.setSquare(file, w(p))          // rank 1
		b.setSquare(8+file, w(Pawn))     // rank 2
		b.setSquare(48+file, Pawn)       // rank 7 (black pawns)
		b.setSquare(56+file, back[file]) // rank 8 (black back rank)
	}

	b.WhiteCanCastleKingSide = true
	b.WhiteCanCastleQueenSide = true
	b.BlackCanCastleKingSide = true
	b.BlackCanCastleQueenSide = true

	return b
}

// deltaIsLegalForPiece checks that a from->to move is geometrically plausible
// for the given piece type on an empty-ish board. This is a coarse guard meant
// to catch garbage moves like the historical "b7f1", not full legality.
func deltaIsLegalForPiece(p Piece, from, to int) bool {
	fromFile, fromRank := from%8, from/8
	toFile, toRank := to%8, to/8
	dFile := toFile - fromFile
	dRank := toRank - fromRank

	abs := func(v int) int {
		if v < 0 {
			return -v
		}
		return v
	}
	af, ar := abs(dFile), abs(dRank)

	switch p.Type() {
	case Pawn:
		// pushes (0 file change, 1-2 ranks) or diagonal captures (1 file, 1 rank)
		if dFile == 0 && ar >= 1 && ar <= 2 {
			return true
		}
		if af == 1 && ar == 1 {
			return true
		}
		return false
	case Knight:
		return (af == 1 && ar == 2) || (af == 2 && ar == 1)
	case Bishop:
		return af == ar && af != 0
	case Rook:
		return (dFile == 0) != (dRank == 0) // exactly one axis moves
	case Queen:
		return (af == ar && af != 0) || (dFile == 0) != (dRank == 0)
	case King:
		// one square any direction, or a two-file castling slide
		if af <= 1 && ar <= 1 && (af != 0 || ar != 0) {
			return true
		}
		if ar == 0 && af == 2 {
			return true
		}
		return false
	}
	return false
}

// TestGeneratedMovesHaveLegalDeltas is the smoke test: from the start position,
// for both sides, every generated move must have a geometrically legal delta
// for the piece sitting on its start square. This directly guards against the
// coordinate-convention regression that produced impossible moves (e.g. b7f1).
func TestGeneratedMovesHaveLegalDeltas(t *testing.T) {
	for _, whiteToMove := range []bool{true, false} {
		b := placeStartingPosition(t)
		b.WhiteToMove = whiteToMove

		moves, err := b.GenerateLegalMovesForPosition()
		if err != nil {
			t.Fatalf("GenerateLegalMovesForPosition error: %v", err)
		}
		if len(moves) == 0 {
			t.Fatalf("expected legal moves from the start position (whiteToMove=%v)", whiteToMove)
		}

		for _, m := range moves {
			from := m.StartSquare()
			to := m.TargetSquare()
			piece := b.MailBox[from]

			if piece.IsEmpty() {
				t.Errorf("move %s starts on an empty square %d", m.ToAlgNot(), from)
				continue
			}
			if piece.IsWhite() != whiteToMove {
				t.Errorf("move %s moves a %v piece on the wrong turn (whiteToMove=%v)",
					m.ToAlgNot(), piece, whiteToMove)
			}
			if !deltaIsLegalForPiece(piece, from, to) {
				t.Errorf("illegal delta for %v: move %s (from %d -> to %d)",
					piece.Type(), m.ToAlgNot(), from, to)
			}
		}
	}
}

// TestStartPositionMoveCount sanity-checks that the opening position yields the
// well-known 20 legal moves for the side to move (16 pawn + 4 knight), ignoring
// check filtering which is a no-op here.
func TestStartPositionMoveCount(t *testing.T) {
	b := placeStartingPosition(t)
	b.WhiteToMove = true

	moves, err := b.GenerateLegalMovesForPosition()
	if err != nil {
		t.Fatalf("GenerateLegalMovesForPosition error: %v", err)
	}
	if len(moves) != 20 {
		names := make([]string, 0, len(moves))
		for _, m := range moves {
			names = append(names, m.ToAlgNot())
		}
		t.Errorf("start position generated %d moves, want 20: %v", len(moves), names)
	}
}
