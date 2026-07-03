package engine

import "testing"

func TestMagicSliderAttacksMatchRayReference(t *testing.T) {
	occupancies := []uint64{
		0,
		0xffffffffffffffff,
		0x0000001818000000,
		0x8100000000000081,
		0x0040201008040200,
		0x0002040810204000,
		NewEngine(true).Occupancy(),
	}

	for sq := 0; sq < 64; sq++ {
		for _, occupancy := range occupancies {
			if got, want := rookAttacks(sq, occupancy), rayAttacks(sq, occupancy, lateralDirections); got != want {
				t.Fatalf("rookAttacks(%d, %#x) = %#x, want %#x", sq, occupancy, got, want)
			}

			if got, want := bishopAttacks(sq, occupancy), rayAttacks(sq, occupancy, diagonalDirections); got != want {
				t.Fatalf("bishopAttacks(%d, %#x) = %#x, want %#x", sq, occupancy, got, want)
			}
		}
	}
}

func TestMagicSliderAttacksMatchAllRelevantBlockers(t *testing.T) {
	for sq := 0; sq < 64; sq++ {
		rookMask := sliderRelevantOccupancyMask(sq, lateralDirections)
		for index := 0; index < 1<<bitsSet(rookMask); index++ {
			occupancy := occupancyFromIndex(index, rookMask)
			if got, want := rookAttacks(sq, occupancy), rayAttacks(sq, occupancy, lateralDirections); got != want {
				t.Fatalf("rookAttacks(%d, blocker index %d) = %#x, want %#x", sq, index, got, want)
			}
		}

		bishopMask := sliderRelevantOccupancyMask(sq, diagonalDirections)
		for index := 0; index < 1<<bitsSet(bishopMask); index++ {
			occupancy := occupancyFromIndex(index, bishopMask)
			if got, want := bishopAttacks(sq, occupancy), rayAttacks(sq, occupancy, diagonalDirections); got != want {
				t.Fatalf("bishopAttacks(%d, blocker index %d) = %#x, want %#x", sq, index, got, want)
			}
		}
	}
}

func bitsSet(mask uint64) int {
	count := 0
	for mask != 0 {
		mask &= mask - 1
		count++
	}
	return count
}
