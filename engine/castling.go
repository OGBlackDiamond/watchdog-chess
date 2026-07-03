package engine

func (e *Engine) CastlePathLegal(kingX, y, rookX int, IsWhite bool) bool {
	rookMask, err := SpaceToMask(rookX, y)
	if err != nil {
		return false
	}

	if IsWhite {
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

	enemyAttackMask, err := e.GenerateAttackMask(!IsWhite)
	if err != nil {
		return false
	}

	for _, x := range []int{kingX, kingX + step, kingX + 2*step} {
		mask, err := SpaceToMask(x, y)
		if err != nil {
			return false
		}

		if enemyAttackMask&mask != 0 {
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
		mask, err := SpaceToMask(x, y)

		if err != nil {
			return false
		}

		if occupancy&mask != 0 {
			return false
		}
	}

	return true
}
