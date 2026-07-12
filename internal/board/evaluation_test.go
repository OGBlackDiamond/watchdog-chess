package board

import "testing"

// walkAndVerifyMaterialPST walks the legal move tree to the given depth and
// asserts at every node that the incrementally maintained Board.MaterialPST
// matches a from-scratch recomputation. This catches any mutation path that
// bypasses setSquare/clearSquare (castling rook shuffle, en passant removal,
// promotion piece swap, captures).
func walkAndVerifyMaterialPST(t *testing.T, b *Board, depth int) {
	t.Helper()

	if got, want := b.MaterialPST, b.ComputeMaterialPST(); got != want {
		t.Fatalf("MaterialPST diverged: incremental=%d, from scratch=%d", got, want)
	}

	if depth == 0 {
		return
	}

	for _, move := range b.GenerateLegalMovesForPosition() {
		child := *b
		if err := child.MakeMove(move); err != nil {
			t.Fatalf("MakeMove(%s) error: %v", move.ToAlgNot(), err)
		}
		walkAndVerifyMaterialPST(t, &child, depth-1)
	}
}

// TestMaterialPSTIncrementalStartPosition verifies incremental eval
// consistency over the opening move tree.
func TestMaterialPSTIncrementalStartPosition(t *testing.T) {
	b := placeStartingPosition(t)
	b.WhiteToMove = true

	walkAndVerifyMaterialPST(t, b, 3)
}

// TestMaterialPSTIncrementalSpecialMoves verifies incremental eval consistency
// in a position where castling, en passant, promotions, and captures are all
// reachable within the walk depth.
func TestMaterialPSTIncrementalSpecialMoves(t *testing.T) {
	b := &Board{}
	w := func(p Piece) Piece { return p + Piece(8) }

	b.setSquare(AlgNotToSpace("e1"), w(King))
	b.setSquare(AlgNotToSpace("a1"), w(Rook))
	b.setSquare(AlgNotToSpace("h1"), w(Rook))
	b.setSquare(AlgNotToSpace("b7"), w(Pawn)) // promotion (and capture-promotion on a8/c8)
	b.setSquare(AlgNotToSpace("e5"), w(Pawn)) // en-passant capturer after ...d7d5

	b.setSquare(AlgNotToSpace("e8"), King)
	b.setSquare(AlgNotToSpace("a8"), Rook) // capture-promotion target
	b.setSquare(AlgNotToSpace("c8"), Knight)
	b.setSquare(AlgNotToSpace("d7"), Pawn) // double push enables en passant

	b.WhiteToMove = true
	b.WhiteCanCastleKingSide = true
	b.WhiteCanCastleQueenSide = true

	walkAndVerifyMaterialPST(t, b, 3)
}

// TestEvaluateSideToMoveSign verifies Evaluate is side-to-move relative while
// MaterialPST stays white-relative.
func TestEvaluateSideToMoveSign(t *testing.T) {
	b := &Board{}
	b.setSquare(AlgNotToSpace("e1"), King+Piece(8))
	b.setSquare(AlgNotToSpace("e8"), King)
	b.setSquare(AlgNotToSpace("d4"), Queen+Piece(8)) // white up a queen

	b.WhiteToMove = true
	whiteView := b.Evaluate()
	b.WhiteToMove = false
	blackView := b.Evaluate()

	if whiteView <= 0 {
		t.Errorf("white to move with extra white queen should evaluate positive, got %d", whiteView)
	}
	if blackView != -whiteView {
		t.Errorf("Evaluate not antisymmetric: white view %d, black view %d", whiteView, blackView)
	}
}
