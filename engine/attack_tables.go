package engine

var knightAttacks [64]uint64
var kingAttacks [64]uint64

var pawnAttacks [2][64]uint64
var pawnAttackers [2][64]uint64

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

const (
	blackIndex = 0
	whiteIndex = 1
)

func init() {
	computeAttackMasks()
	initSliderMagics()
}

// computeAttackMasks generates lookup tables for how kings and knights move
// from any square on an otherwise empty board.
func computeAttackMasks() {
	for sq := range 64 {
		knightAttacks[sq] = buildOffsetAttackMask(sq, knightOffsets)
		kingAttacks[sq] = buildOffsetAttackMask(sq, kingOffsets)
	
		pawnAttacks[blackIndex][sq] = buildPawnAttacksFrom(sq, false)
		pawnAttacks[whiteIndex][sq] = buildPawnAttacksFrom(sq, true)
		
		pawnAttackers[blackIndex][sq] = buildPawnAttackersTo(sq, false)
		pawnAttackers[whiteIndex][sq] = buildPawnAttackersTo(sq, true)
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

		mask, ok := SpaceToMask(toX, toY)

		if !ok {
			panic("Space out of bounds")
		}

		attacks |= mask

	}

	return attacks
}

func buildPawnAttackersTo(targetSq int, byWhite bool) uint64 {
    x, y, err := MaskToSpace(uint64(1) << targetSq)
    if err != nil {
        panic(err)
    }

    pawnY := y + 1
    if !byWhite {
        pawnY = y - 1
    }

    attackers := uint64(0)

    for _, dx := range [2]int{-1, 1} {
        pawnX := x + dx

        if CheckBounds(pawnX, pawnY) {
            continue
        }

        mask, ok := SpaceToMask(pawnX, pawnY)
        if !ok {
            panic("Space out of bounds")
        }

        attackers |= mask
    }

    return attackers
}

func buildPawnAttacksFrom(fromSq int, byWhite bool) uint64 {
	x, y, err := MaskToSpace(uint64(1) << fromSq)
	if err != nil {
		panic(err)
	}

	direction := 1
	if byWhite {
		direction = -1
	}

	attacks := uint64(0)

	for _, dx := range [2]int{-1, 1} {
        toX := x + dx
        toY := y + direction

        if CheckBounds(toX, toY) {
            continue
        }

        mask, ok := SpaceToMask(toX, toY)
        if !ok {
            panic("Space out of bounds")
        }

        attacks |= mask
    }

	return attacks
}
