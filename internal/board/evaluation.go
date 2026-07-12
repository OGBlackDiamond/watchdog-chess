package board

const (
	pawnValue   int = 100
	knightValue int = 320
	bishopValue int = 330
	rookValue   int = 500
	queenValue  int = 900
	kingValue   int = 20_000
)

// pieceValues and pieceTables are indexed directly by Piece.Type()
// (Pawn=1, Rook=2, Knight=3, Bishop=4, Queen=5, King=6). Index 0 is NONE.
var (
	pieceValues = [7]int{
		0, // NONE
		pawnValue,
		rookValue,
		knightValue,
		bishopValue,
		queenValue,
		kingValue,
	}

	pieceTables = [7][]int{
		nil, // NONE
		pawnTableMiddle[:],
		rookTableMiddle[:],
		knightTableMiddle[:],
		bishopTableMiddle[:],
		queenTableMiddle[:],
		kingTableMiddle[:],
	}
)

// Evaluate returns the score of the position relative to the side to move.
func (b *Board) Evaluate() int {
	if b.WhiteToMove {
		return b.MaterialPST
	}
	return -b.MaterialPST
}

// PieceValue returns the material value of a piece (color-independent).
func PieceValue(piece Piece) int {
	return pieceValues[piece.Type()]
}

func pieceSquareValue(piece Piece, square int) int {
	return pieceTables[piece.Type()][pstIndex(square, piece.IsWhite())]
}

// ComputeMaterialPST recalculates the white-relative material + piece-square
// score from scratch. It is used to initialize Board.MaterialPST after setting
// up a position, and by tests to verify the incremental updates performed in
// setSquare/clearSquare.
//
// Note: this is deliberately NOT side-to-move relative - the accumulator is
// always from white's perspective; Evaluate applies the side-to-move flip.
func (b *Board) ComputeMaterialPST() int {
	score := 0

	for square, piece := range b.MailBox {
		if piece.Type() == NONE {
			continue
		}

		v := PieceValue(piece) + pieceSquareValue(piece, square)

		if piece.IsWhite() {
			score += v
		} else {
			score -= v
		}
	}

	return score
}
