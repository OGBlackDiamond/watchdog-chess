package engine

func (e *Engine) CastlePathLegal(kingX, y, rookX int, isWhite bool) bool {
	rookMask, ok := SpaceToMask(rookX, y)
	if !ok {
		return false
	}

	if isWhite {
		if e.Board.WhitePieces.Rooks&rookMask == 0 {
			return false
		}
	} else {
		if e.Board.BlackPieces.Rooks&rookMask == 0 {
			return false
		}
	}

	if !e.CastlePathClear(kingX, y, rookX) {
		return false
	}

	step := 1
	if rookX < kingX {
		step = -1
	}

	for _, x := range []int{kingX, kingX + step, kingX + 2*step} {
		mask, ok := SpaceToMask(x, y)
		if !ok {
			return false
		}

		if e.SquareIsAttackedBy(mask, !isWhite) {
			return false
		}
	}

	return true
}

func (e *Engine) CastlePathClear(kingX, y, rookX int) bool {
	step := 1

	if rookX < kingX {
		step = -1
	}

	occupancy := e.Occupancy()

	for x := kingX + step; x != rookX; x += step {
		mask, ok := SpaceToMask(x, y)

		if !ok {
			return false
		}

		if occupancy&mask != 0 {
			return false
		}
	}

	return true
}
