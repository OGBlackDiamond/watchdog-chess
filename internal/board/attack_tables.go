package board

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

	file, rank, err := MaskToGrid(uint64(1) << sq)

	if err != nil {
		panic(err)
	}

	attacks := uint64(0)

	for _, offset := range offsets {
		toFile := file + offset[0]
		toRank := rank + offset[1]

		if !onBoard(toFile, toRank) {
			continue
		}

		mask, ok := GridToMask(toFile, toRank)

		if !ok {
			panic("Space out of bounds")
		}

		attacks |= mask

	}

	return attacks
}

func buildPawnAttackersTo(targetSq int, byWhite bool) uint64 {
	file, rank, err := MaskToGrid(uint64(1) << targetSq)
	if err != nil {
		panic(err)
	}

	// A white pawn attacks upward, so an attacker of targetSq sits one rank
	// below it; a black pawn attacks downward, so its attacker sits one rank
	// above the target.
	pawnRank := rank - 1
	if !byWhite {
		pawnRank = rank + 1
	}

	attackers := uint64(0)

	for _, dFile := range [2]int{-1, 1} {
		pawnFile := file + dFile

		if !onBoard(pawnFile, pawnRank) {
			continue
		}

		mask, ok := GridToMask(pawnFile, pawnRank)
		if !ok {
			panic("Space out of bounds")
		}

		attackers |= mask
	}

	return attackers
}

func buildPawnAttacksFrom(fromSq int, byWhite bool) uint64 {
	file, rank, err := MaskToGrid(uint64(1) << fromSq)
	if err != nil {
		panic(err)
	}

	// White pawns advance toward higher ranks, black toward lower ranks.
	direction := -1
	if byWhite {
		direction = 1
	}

	attacks := uint64(0)

	for _, dFile := range [2]int{-1, 1} {
		toFile := file + dFile
		toRank := rank + direction

		if !onBoard(toFile, toRank) {
			continue
		}

		mask, ok := GridToMask(toFile, toRank)
		if !ok {
			panic("Space out of bounds")
		}

		attacks |= mask
	}

	return attacks
}
