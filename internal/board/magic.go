package board

import "math/bits"

// Magic bitboard tables for sliding-piece attacks.

type magicEntry struct {
	mask    uint64
	magic   uint64
	shift   uint
	attacks []uint64
}

var (
	rookMagics   [64]magicEntry
	bishopMagics [64]magicEntry
)

var rookDirs = [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
var bishopDirs = [4][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

// slidingAttacks computes attacks from sq given a blocker set, sliding along dirs.
func slidingAttacks(sq int, occ uint64, dirs [4][2]int) uint64 {
	var attacks uint64
	f0, r0 := fileOf(sq), rankOf(sq)
	for _, d := range dirs {
		f, r := f0+d[0], r0+d[1]
		for onBoard(f, r) {
			s := r*8 + f
			attacks = setBit(attacks, s)
			if testBit(occ, s) {
				break
			}
			f += d[0]
			r += d[1]
		}
	}
	return attacks
}

// sliderMask computes the relevant occupancy mask (excluding board edges).
func sliderMask(sq int, rook bool) uint64 {
	var mask uint64
	f0, r0 := fileOf(sq), rankOf(sq)
	if rook {
		for r := r0 + 1; r <= 6; r++ {
			mask = setBit(mask, r*8+f0)
		}
		for r := r0 - 1; r >= 1; r-- {
			mask = setBit(mask, r*8+f0)
		}
		for f := f0 + 1; f <= 6; f++ {
			mask = setBit(mask, r0*8+f)
		}
		for f := f0 - 1; f >= 1; f-- {
			mask = setBit(mask, r0*8+f)
		}
	} else {
		for f, r := f0+1, r0+1; f <= 6 && r <= 6; f, r = f+1, r+1 {
			mask = setBit(mask, r*8+f)
		}
		for f, r := f0+1, r0-1; f <= 6 && r >= 1; f, r = f+1, r-1 {
			mask = setBit(mask, r*8+f)
		}
		for f, r := f0-1, r0+1; f >= 1 && r <= 6; f, r = f-1, r+1 {
			mask = setBit(mask, r*8+f)
		}
		for f, r := f0-1, r0-1; f >= 1 && r >= 1; f, r = f-1, r-1 {
			mask = setBit(mask, r*8+f)
		}
	}
	return mask
}

// xorshift PRNG for magic candidate generation (deterministic seed).
type rng struct{ state uint64 }

func (r *rng) next() uint64 {
	r.state ^= r.state >> 12
	r.state ^= r.state << 25
	r.state ^= r.state >> 27
	return r.state * 2685821657736338717
}

func (r *rng) sparse() uint64 {
	return r.next() & r.next() & r.next()
}

func findMagics(rook bool, table *[64]magicEntry) {
	r := rng{state: 0x1234567890abcdef}
	if !rook {
		r.state = 0xfedcba0987654321
	}
	var dirs [4][2]int
	if rook {
		dirs = rookDirs
	} else {
		dirs = bishopDirs
	}
	for sq := 0; sq < 64; sq++ {
		mask := sliderMask(sq, rook)
		n := bits.OnesCount64(mask)
		size := 1 << uint(n)
		occs := make([]uint64, size)
		refs := make([]uint64, size)

		// enumerate all subsets of mask via carry-rippler.
		var occ uint64
		i := 0
		for {
			occs[i] = occ
			refs[i] = slidingAttacks(sq, occ, dirs)
			i++
			occ = (occ - mask) & mask
			if occ == 0 {
				break
			}
		}

		shift := uint(64 - n)
		used := make([]uint64, size)
		seen := make([]int, size)
		stamp := 0
		for {
			magic := r.sparse()
			// heuristic reject: need enough high bits spread.
			if bits.OnesCount64((mask*magic)&0xFF00000000000000) < 6 {
				continue
			}
			stamp++
			fail := false
			for k := 0; k < size; k++ {
				idx := uint((occs[k] * magic) >> shift)
				if seen[idx] != stamp {
					seen[idx] = stamp
					used[idx] = refs[k]
				} else if used[idx] != refs[k] {
					fail = true
					break
				}
			}
			if !fail {
				attacks := make([]uint64, size)
				for k := 0; k < size; k++ {
					idx := uint((occs[k] * magic) >> shift)
					attacks[idx] = refs[k]
				}
				table[sq] = magicEntry{mask: mask, magic: magic, shift: shift, attacks: attacks}
				break
			}
		}
	}
}

func rookAttacks(sq int, occ uint64) uint64 {
	m := &rookMagics[sq]
	return m.attacks[((occ&m.mask)*m.magic)>>m.shift]
}

func bishopAttacks(sq int, occ uint64) uint64 {
	m := &bishopMagics[sq]
	return m.attacks[((occ&m.mask)*m.magic)>>m.shift]
}

func queenAttacks(sq int, occ uint64) uint64 {
	return rookAttacks(sq, occ) | bishopAttacks(sq, occ)
}

func init() {
	findMagics(true, &rookMagics)
	findMagics(false, &bishopMagics)
}
