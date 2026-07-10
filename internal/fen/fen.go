// Package fen interprets FEN strings into a format that is readable for the engine
package fen

import (
	"fmt"
	"strings"

	"github.com/OGBlackDiamond/watchdog-chess/internal/board"
)

// black pieces are lowercase

// StartingPositionFEN is the FEN string for the standard chess starting position.
const StartingPositionFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"


func NewBoardFromFen(fen string) (*board.Board, error) {

	fields := strings.Fields(strings.TrimSpace(fen))

	if len(fields) < 4 {
		return nil, fmt.Errorf("fen %q: expected at least 4 fields, got %d", fen, len(fields))
	}

	b := &board.Board{}

	ranks := strings.Split(fields[0], "/")
	if len(ranks) != 8 {
		return nil, fmt.Errorf("fen %q: expected 8 ranks, got %d", fen, len(ranks))
	}

	// field 0, board state
	for i, rankStr := range ranks {
		// first FEN rank is rank 8
		rank := 7 - i
		file := 0

		for _, c := range rankStr {
			
			// spaces
			if c >= '1' && c <= '8' {
				file += int(c - '0')
				continue
			}

			if file > 7 {
				return nil, fmt.Errorf("fen %q: rank %d overflows 8 files", fen, rank+1)
			}

			square := rank*8 + file
			mask := uint64(1) << square

			pieces := &b.BlackPieces
			colorMask := 0b0000
			if c >= 'A' && c <= 'Z' {
				// white pieces are uppercase and carry the color bit (0b1000)
				pieces = &b.WhitePieces
				colorMask = 0b1000
			}

			// interesting switch syntax to catch upper and lower
			switch c {
			case 'p', 'P':
				pieces.Pawns |= mask
				b.MailBox[square] = board.Pawn + board.Piece(colorMask)
			case 'r', 'R':
				pieces.Rooks |= mask
				b.MailBox[square] = board.Rook + board.Piece(colorMask)
			case 'n', 'N':
				pieces.Knights |= mask
				b.MailBox[square] = board.Knight + board.Piece(colorMask)
			case 'b', 'B':
				pieces.Bishops |= mask
				b.MailBox[square] = board.Bishop + board.Piece(colorMask)
			case 'q', 'Q':
				pieces.Queen |= mask
				b.MailBox[square] = board.Queen + board.Piece(colorMask)
			case 'k', 'K':
				pieces.King |= mask
				b.MailBox[square] = board.King + board.Piece(colorMask)
			default:
				return nil, fmt.Errorf("fen %q: invalid piece character %q", fen, c)
			}

			file++

		}

		if file != 8 {
			return nil, fmt.Errorf("fen %q: rank %d has %d files, expected 8", fen, rank+1, file)
		}
	}


	// field 1, side to move
	switch fields[1] {
	case "w":
		b.WhiteToMove = true
	case "b":
		b.WhiteToMove = false
	default:
		return nil, fmt.Errorf("fen %q: invalid side to move %q", fen, fields[1])
	}

	// field 2, castling rights
	if fields[2] != "-" {
		for _, c := range fields[2] {
			switch c {
			case 'K':
				b.WhiteCanCastleKingSide = true
			case 'Q':
				b.WhiteCanCastleQueenSide = true
			case 'k':
				b.BlackCanCastleKingSide = true
			case 'q':
				b.BlackCanCastleQueenSide = true
			default:
				return nil, fmt.Errorf("fen %q: invalid castling character %q", fen, c)
			}
		}
	}

	// field 3: en passant target square
	if fields[3] != "-" {
		sq := fields[3]

		if len(sq) != 2 || sq[0] < 'a' || sq[0] > 'h' || (sq[1] != '3' && sq[1] != '6') {
			return nil, fmt.Errorf("fen %q: invalid en passant square %q", fen, sq)
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

		b.EnPassantTarget = uint64(1) << (rank*8 + file)
		b.EnPassantPieceMask = uint64(1) << (pawnRank*8 + file)
	}
	

	return b, nil

}
