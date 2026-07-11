package perft

import (
	"strconv"
	"testing"

	"github.com/OGBlackDiamond/watchdog-chess/internal/fen"
)

// perftCase is a single (position, depth) expectation drawn from the standard
// perft reference values documented on the Chess Programming Wiki.
type perftCase struct {
	depth int
	nodes uint64
	// long marks deep counts that are skipped under `go test -short`.
	long bool
}

type perftSuite struct {
	name  string
	fen   string
	cases []perftCase
}

// The canonical perft test positions and their known-correct node counts.
var suites = []perftSuite{
	{
		name: "startpos",
		fen:  fen.StartingPositionFEN,
		cases: []perftCase{
			{1, 20, false},
			{2, 400, false},
			{3, 8902, false},
			{4, 197281, false},
			{5, 4865609, false},
			{6, 119060324, true},
		},
	},
	{
		name: "kiwipete",
		fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		cases: []perftCase{
			{1, 48, false},
			{2, 2039, false},
			{3, 97862, false},
			{4, 4085603, false},
			{5, 193690690, true},
		},
	},
	{
		name: "position3",
		fen:  "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		cases: []perftCase{
			{1, 14, false},
			{2, 191, false},
			{3, 2812, false},
			{4, 43238, false},
			{5, 674624, false},
			{6, 11030083, true},
		},
	},
	{
		name: "position4",
		fen:  "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		cases: []perftCase{
			{1, 6, false},
			{2, 264, false},
			{3, 9467, false},
			{4, 422333, false},
			{5, 15833292, true},
		},
	},
	{
		name: "position5",
		fen:  "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		cases: []perftCase{
			{1, 44, false},
			{2, 1486, false},
			{3, 62379, false},
			{4, 2103487, true},
		},
	},
	{
		name: "position6",
		fen:  "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		cases: []perftCase{
			{1, 46, false},
			{2, 2079, false},
			{3, 89890, false},
			{4, 3894594, true},
		},
	},
}

func TestPerft(t *testing.T) {
	for _, s := range suites {
		s := s
		t.Run(s.name, func(t *testing.T) {
			for _, c := range s.cases {
				c := c
				if c.long && testing.Short() {
					continue
				}

				name := "depth" + strconv.Itoa(c.depth)
				t.Run(name, func(t *testing.T) {
					b, err := fen.NewBoardFromFen(s.fen)
					if err != nil {
						t.Fatalf("NewBoardFromFen(%q) error: %v", s.fen, err)
					}

					got, err := Perft(b, c.depth)
					if err != nil {
						t.Fatalf("Perft(%s, %d) error: %v", s.name, c.depth, err)
					}
					if got != c.nodes {
						t.Errorf("Perft(%s, depth %d) = %d, want %d (diff %+d)",
							s.name, c.depth, got, c.nodes, int64(got)-int64(c.nodes))
					}
				})
			}
		})
	}
}

// TestDivideMatchesPerft verifies that the sum of the per-move divide counts
// equals the plain perft count for the same depth.
func TestDivideMatchesPerft(t *testing.T) {
	b, err := fen.NewBoardFromFen(fen.StartingPositionFEN)
	if err != nil {
		t.Fatal(err)
	}

	const depth = 3
	entries, total, err := Divide(b, depth)
	if err != nil {
		t.Fatalf("Divide error: %v", err)
	}

	want, err := Perft(b, depth)
	if err != nil {
		t.Fatalf("Perft error: %v", err)
	}
	if total != want {
		t.Errorf("Divide total = %d, Perft = %d", total, want)
	}

	var sum uint64
	for _, e := range entries {
		sum += e.Nodes
	}
	if sum != total {
		t.Errorf("sum of divide entries = %d, reported total = %d", sum, total)
	}
}
