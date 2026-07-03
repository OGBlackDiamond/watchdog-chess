package engine

import (
	"fmt"
	"strings"
)

// StartingPositionFEN is the FEN string for the standard chess starting position.
const StartingPositionFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// NewEngineFromFEN builds an Engine from a FEN string
// (https://www.chessprogramming.org/Forsyth-Edwards_Notation).
//
// The returned engine always uses the standard orientation (PlayAsWhite=true,
// white on the bottom rows). With that orientation the internal bit layout
// matches little-endian rank-file mapping: bit 0 = a1, bit 63 = h8.
//
// The engine does not track whose turn it is, so the side to move is returned
// separately. The halfmove clock and fullmove number fields are ignored
// because the engine does not track them.
func NewEngineFromFEN(fen string) (*Engine, bool, error) {
	fields := strings.Fields(strings.TrimSpace(fen))

	if len(fields) < 4 {
		return nil, false, fmt.Errorf("fen %q: expected at least 4 fields, got %d", fen, len(fields))
	}

	e := &Engine{PlayAsWhite: true}

	// field 0: piece placement, ranks 8 down to 1
	ranks := strings.Split(fields[0], "/")
	if len(ranks) != 8 {
		return nil, false, fmt.Errorf("fen %q: expected 8 ranks, got %d", fen, len(ranks))
	}

	for i, rankStr := range ranks {
		rank := 7 - i // 0-based rank index: first FEN rank is rank 8
		file := 0

		for _, c := range rankStr {
			if c >= '1' && c <= '8' {
				file += int(c - '0')
				continue
			}

			if file > 7 {
				return nil, false, fmt.Errorf("fen %q: rank %d overflows 8 files", fen, rank+1)
			}

			mask := uint64(1) << (rank*8 + file)

			pieces := &e.Board.BlackPieces
			if c >= 'A' && c <= 'Z' {
				pieces = &e.Board.WhitePieces
			}

			switch c {
			case 'p', 'P':
				pieces.Pawns |= mask
			case 'r', 'R':
				pieces.Rooks |= mask
			case 'n', 'N':
				pieces.Knights |= mask
			case 'b', 'B':
				pieces.Bishops |= mask
			case 'q', 'Q':
				pieces.Queen |= mask
			case 'k', 'K':
				pieces.King |= mask
			default:
				return nil, false, fmt.Errorf("fen %q: invalid piece character %q", fen, c)
			}

			file++
		}

		if file != 8 {
			return nil, false, fmt.Errorf("fen %q: rank %d has %d files, expected 8", fen, rank+1, file)
		}
	}

	// field 1: side to move
	var whiteToMove bool
	switch fields[1] {
	case "w":
		whiteToMove = true
	case "b":
		whiteToMove = false
	default:
		return nil, false, fmt.Errorf("fen %q: invalid side to move %q", fen, fields[1])
	}

	// field 2: castling rights
	if fields[2] != "-" {
		for _, c := range fields[2] {
			switch c {
			case 'K':
				e.whiteCanCastleKingSide = true
			case 'Q':
				e.whiteCanCastleQueenSide = true
			case 'k':
				e.blackCanCastleKingSide = true
			case 'q':
				e.blackCanCastleQueenSide = true
			default:
				return nil, false, fmt.Errorf("fen %q: invalid castling character %q", fen, c)
			}
		}
	}

	// field 3: en passant target square
	if fields[3] != "-" {
		sq := fields[3]

		if len(sq) != 2 || sq[0] < 'a' || sq[0] > 'h' || (sq[1] != '3' && sq[1] != '6') {
			return nil, false, fmt.Errorf("fen %q: invalid en passant square %q", fen, sq)
		}

		file := int(sq[0] - 'a')
		rank := int(sq[1] - '1') // 0-based: 2 (rank 3) or 5 (rank 6)

		// the pawn that just double-pushed sits one rank beyond the target
		// square: target on rank 3 -> white pawn on rank 4, target on
		// rank 6 -> black pawn on rank 5
		pawnRank := 3
		if rank == 5 {
			pawnRank = 4
		}

		e.enPassantTarget = uint64(1) << (rank*8 + file)
		e.enPassantPieceMask = uint64(1) << (pawnRank*8 + file)
	}

	return e, whiteToMove, nil
}
