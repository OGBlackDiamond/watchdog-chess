package engine

var knightAttacks [64]uint64
var kingAttacks [64]uint64

var knightOffsets = [][2]int{
	{1, 2},
	{-1, 2},
	{-2, 1},
	{-2, -1},
	{1, -2},
	{-1, -2},
	{2, -1},
	{2, 1},
}

var kingOffsets = [][2]int{
	{0, 1},
	{-1, 1},
	{-1, 0},
	{-1, -1},
	{0, -1},
	{1, -1},
	{1, 0},
	{1, 1},
}

func init() {
	computeAttackMasks()
	initSliderMagics()
}

// computeAttackMasks generates lookup tables for how kings and knights move
// from any square on an otherwise empty board.
func computeAttackMasks() {
	for sq := 0; sq < 64; sq++ {
		knightAttacks[sq] = buildOffsetAttackMask(sq, knightOffsets)
		kingAttacks[sq] = buildOffsetAttackMask(sq, kingOffsets)
	}
}

func buildOffsetAttackMask(sq int, offsets [][2]int) uint64 {

	x, y, err := MaskToSpace(uint64(1) << sq)

	if err != nil {
		panic(err)
	}

	attacks := uint64(0)

	for _, offset := range offsets {
		toX := x + offset[0]
		toY := y + offset[1]

		if CheckBounds(toX, toY) {
			continue
		}

		mask, err := SpaceToMask(toX, toY)

		if err != nil {
			panic(err)
		}

		attacks |= mask

	}

	return attacks
}
