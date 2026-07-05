package watchdog

import "github.com/OGBlackDiamond/watchdog-chess/engine"

type TTEntry struct {
	hash     uint64
	depth    int
	score    float64
	flag     BoundFlag
	bestMove engine.Move
}

type BoundFlag int

const (
	Exact BoundFlag = iota
	LowerBound
	UpperBound
)
