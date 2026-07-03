package watchdog

import "github.com/OGBlackDiamond/watchdog-chess/engine"

func Evaluate(board engine.Board, whiteToMove bool) (float64, error) {

	whiteScore, _ := material(board.WhitePieces)
	blackScore, _ := material(board.BlackPieces)

	score := whiteScore - blackScore
	if !whiteToMove {
		score *= -1
	}

	return float64(score), nil
}

func material(pieces engine.Pieces) (int, error) {
	material := []uint64{
		pieces.Pawns,
		pieces.Rooks,
		pieces.Knights,
		pieces.Bishops,
		pieces.Queen,
	}

	total := 0

	for piece, mask := range material {

		pieceTotal := 0

		for i := range 64 {
			if mask>>i%2 != 0 {
				pieceTotal++
			}
		}

		switch engine.Piece(piece) {
		case engine.Pawn:
			total += pieceTotal * 100

		case engine.Rook:
			total += pieceTotal * 500

		case engine.Knight:
			fallthrough // this is sick
		case engine.Bishop:
			total += pieceTotal * 300

		case engine.Queen:
			total += pieceTotal * 900

		}
	}

	return total, nil
}
