package engine

import (
	"math/bits"
	"sync/atomic"
	"unsafe"

	"github.com/OGBlackDiamond/watchdog-chess/internal/board"
)


type TT struct {
	entries []TTEntry
	mask uint64
}

type TTEntry struct {
	key atomic.Uint64 // position hash
	data atomic.Uint64 // information about the positon
}

type Bound uint8
const (
    BoundNone  Bound = 0  // empty slot / no info
    BoundExact Bound = 1
    BoundLower Bound = 2
    BoundUpper Bound = 3
)

// Resize resizes the given table to a size in mb - this must be a power of two
func (tt *TT) Resize(mb int) {
    n := entriesForMB(mb) // must be a power of two
    tt.entries = make([]TTEntry, n)
    tt.mask = n - 1
}


// Bit order is as follows
// MSB ---
// unused -> 11 bits
// score  -> 21 bits
// bound  -> 2 bits
// depth  -> 8 bits
// move   -> 16 bits
func pack(m board.Move, depth, score, bound int) uint64 {
    return uint64(uint16(m)) |
        uint64(depth&0xFF)<<16 |
        uint64(bound&0x3)<<24 |
        uint64((score+Infinity)&0x1FFFFF)<<32 // 0x1FFFFF = 21 bits
}

func unpack(d uint64) (m board.Move, depth, score, bound int) {
    m = board.Move(d)                    // low 16 bits, no shift
    depth = int(d>>16) & 0xFF
    bound = int(d>>24) & 0x3
    score = int((d>>32)&0x1FFFFF) - Infinity
    return
}

func entriesForMB(mb int) uint64 {
    bytes := uint64(mb) * 1024 * 1024
    entrySize := uint64(unsafe.Sizeof(TTEntry{}))
    count := max(bytes / entrySize, 1)
    // round DOWN to a power of two so index = key & (count-1)
    return uint64(1) << (bits.Len64(count) - 1)
}
