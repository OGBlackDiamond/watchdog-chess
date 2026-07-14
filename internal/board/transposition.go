package board

import (
	"math/bits"
	"sync/atomic"
	"unsafe"
)


type TT []TTEntry

type TTEntry struct {
	key atomic.Uint64 // zobrist value ^ position
	data atomic.Uint64 // information about the positon
}


func entriesForMB(mb int) uint64 {
    bytes := uint64(mb) * 1024 * 1024
    entrySize := uint64(unsafe.Sizeof(TTEntry{})) // 16 for two atomic.Uint64
    count := bytes / entrySize
    if count < 1 {
        count = 1
    }
    // round DOWN to a power of two so index = key & (count-1)
    return uint64(1) << (bits.Len64(count) - 1)
}
