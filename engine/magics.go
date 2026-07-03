package engine

import "math/bits"

type sliderMagic struct {
	mask    uint64
	magic   uint64
	shift   uint
	attacks []uint64
}

var rookMagics [64]sliderMagic
var bishopMagics [64]sliderMagic

var lateralDirections = [][2]int{
	{1, 0},
	{-1, 0},
	{0, 1},
	{0, -1},
}

var diagonalDirections = [][2]int{
	{1, 1},
	{-1, 1},
	{1, -1},
	{-1, -1},
}

func initSliderMagics() {
	for sq := 0; sq < 64; sq++ {
		rookMagics[sq] = buildSliderMagic(sq, lateralDirections)
		bishopMagics[sq] = buildSliderMagic(sq, diagonalDirections)
	}
}

func rookAttacks(square int, occupancy uint64) uint64 {
	entry := rookMagics[square]
	index := ((occupancy & entry.mask) * entry.magic) >> entry.shift
	return entry.attacks[index]
}

func bishopAttacks(square int, occupancy uint64) uint64 {
	entry := bishopMagics[square]
	index := ((occupancy & entry.mask) * entry.magic) >> entry.shift
	return entry.attacks[index]
}

func buildSliderMagic(square int, directions [][2]int) sliderMagic {
	mask := sliderRelevantOccupancyMask(square, directions)
	relevantBits := bits.OnesCount64(mask)
	tableSize := 1 << relevantBits

	occupancies := make([]uint64, tableSize)
	attacks := make([]uint64, tableSize)

	for index := range tableSize {
		occupancy := occupancyFromIndex(index, mask)
		occupancies[index] = occupancy
		attacks[index] = rayAttacks(square, occupancy, directions)
	}

	magic := findMagic(square, mask, relevantBits, occupancies, attacks)
	entry := sliderMagic{
		mask:    mask,
		magic:   magic,
		shift:   uint(64 - relevantBits),
		attacks: make([]uint64, tableSize),
	}

	for index, occupancy := range occupancies {
		magicIndex := ((occupancy & entry.mask) * entry.magic) >> entry.shift
		entry.attacks[magicIndex] = attacks[index]
	}

	return entry
}

func findMagic(square int, mask uint64, relevantBits int, occupancies []uint64, attacks []uint64) uint64 {
	tableSize := 1 << relevantBits
	usedAttacks := make([]uint64, tableSize)
	usedEpochs := make([]uint32, tableSize)
	epoch := uint32(0)
	rng := xorshift64{state: 0x9e3779b97f4a7c15 ^ uint64(square+1)*0xbf58476d1ce4e5b9 ^ uint64(relevantBits)*0x94d049bb133111eb}
	shift := uint(64 - relevantBits)

	for attempts := 0; attempts < 10_000_000; attempts++ {
		magic := rng.sparse()
		if bits.OnesCount64((mask*magic)&0xff00000000000000) < 6 {
			continue
		}

		epoch++
		if epoch == 0 {
			clear(usedEpochs)
			epoch = 1
		}

		failed := false
		for index, occupancy := range occupancies {
			magicIndex := (occupancy * magic) >> shift
			if usedEpochs[magicIndex] != epoch {
				usedEpochs[magicIndex] = epoch
				usedAttacks[magicIndex] = attacks[index]
				continue
			}

			if usedAttacks[magicIndex] != attacks[index] {
				failed = true
				break
			}
		}

		if !failed {
			return magic
		}
	}

	panic("failed to find slider magic")
}

func sliderRelevantOccupancyMask(square int, directions [][2]int) uint64 {
	x, y := squareToCoords(square)
	mask := uint64(0)

	for _, dir := range directions {
		file := x + dir[0]
		rank := y + dir[1]

		for !CheckBounds(file, rank) {
			if CheckBounds(file+dir[0], rank+dir[1]) {
				break
			}

			mask |= mustSpaceToMask(file, rank)
			file += dir[0]
			rank += dir[1]
		}
	}

	return mask
}

func rayAttacks(square int, occupancy uint64, directions [][2]int) uint64 {
	x, y := squareToCoords(square)
	attacks := uint64(0)

	for _, dir := range directions {
		file := x + dir[0]
		rank := y + dir[1]

		for !CheckBounds(file, rank) {
			mask := mustSpaceToMask(file, rank)
			attacks |= mask

			if occupancy&mask != 0 {
				break
			}

			file += dir[0]
			rank += dir[1]
		}
	}

	return attacks
}

func occupancyFromIndex(index int, mask uint64) uint64 {
	occupancy := uint64(0)

	for bit := 0; mask != 0; bit++ {
		squareMask := mask & -mask
		mask &= mask - 1

		if index&(1<<bit) != 0 {
			occupancy |= squareMask
		}
	}

	return occupancy
}

func squareToCoords(square int) (int, int) {
	return square % 8, 7 - square/8
}

func mustSpaceToMask(x, y int) uint64 {
	mask, err := SpaceToMask(x, y)
	if err != nil {
		panic(err)
	}

	return mask
}

type xorshift64 struct {
	state uint64
}

func (r *xorshift64) next() uint64 {
	r.state ^= r.state << 13
	r.state ^= r.state >> 7
	r.state ^= r.state << 17
	return r.state
}

func (r *xorshift64) sparse() uint64 {
	return r.next() & r.next() & r.next()
}
