package engine

import (
	"math/bits"
	"sync/atomic"
	"unsafe"

	"github.com/OGBlackDiamond/watchdog-chess/internal/board"
)

type TT struct {
	entries    []TTEntry
	mask       uint64
	generation int
}

type TTEntry struct {
	key  atomic.Uint64 // position hash
	data atomic.Uint64 // information about the positon
}

// Bit order for data is as follows
// MSB ---
// unused -> 11 bits
// score  -> 21 bits
// age    -> 6 bits
// bound  -> 2 bits
// depth  -> 8 bits
// move   -> 16 bits

type Bound uint8

const (
	BoundNone  Bound = 0 // empty slot / no info
	BoundExact Bound = 1
	BoundLower Bound = 2
	BoundUpper Bound = 3
)

func (tt *TT) getEntry(key uint64) (m board.Move, depth, score, age, bound int, hit bool) {
	if tt == nil || len(tt.entries) == 0 {
		return 0, 0, 0, 0, 0, false
	}

	e := &tt.entries[key&tt.mask]
	k := e.key.Load()
	d := e.data.Load()
	if d == 0 || k^d != key { // test the full key instead of the low bits on the index
		return 0, 0, 0, 0, 0, false
	}
	m, depth, score, age, bound = unpack(d)
	return m, depth, score, age, bound, true
}

func (tt *TT) storeEntry(key uint64, m board.Move, depth, score, bound int) {
	if tt == nil || len(tt.entries) == 0 {
		return
	}

	d := pack(m, depth, score, tt.generation, bound)
	e := &tt.entries[key&tt.mask]
	e.key.Store(key ^ d)
	e.data.Store(d)
}

func pack(m board.Move, depth, score, age, bound int) uint64 {
	return uint64(uint16(m)) |
		uint64(depth&0xFF)<<16 |
		uint64(bound&0x3)<<24 |
		uint64(age&0x3F)<<30 |
		uint64((score+Infinity)&0x1FFFFF)<<38 // 0x1FFFFF = 21 bits
}

func unpack(d uint64) (m board.Move, depth, score, age, bound int) {
	m = board.Move(d) // low 16 bits, no shift
	depth = int(d>>16) & 0xFF
	bound = int(d>>24) & 0x3
	age = int(d>>30) & 0x3F
	score = int((d>>38)&0x1FFFFF) - Infinity
	return
}

// Resize resizes the given table to a size in mb - this must be a power of two
func (tt *TT) Resize(mb int) {
	n := entriesForMB(mb) // must be a power of two
	tt.entries = make([]TTEntry, n)
	tt.mask = n - 1
}

func entriesForMB(mb int) uint64 {
	bytes := uint64(mb) * 1024 * 1024
	entrySize := uint64(unsafe.Sizeof(TTEntry{}))
	count := max(bytes/entrySize, 1)
	// round DOWN to a power of two so index = key & (count-1)
	return uint64(1) << (bits.Len64(count) - 1)
}

// normalize mate distance for node relative
func valueToTT(v, ply int) int {
	if v >= mateScore-MaxPly {
		return v + ply
	}
	if v <= -(mateScore - MaxPly) {
		return v - ply
	}
	return v
}

func valueFromTT(v, ply int) int {
	if v >= mateScore-MaxPly {
		return v - ply
	}
	if v <= -(mateScore - MaxPly) {
		return v + ply
	}
	return v
}
