# watchdog-chess

A chess engine I made. An Ebiten GUI where you play against Watchdog, a
negamax/alpha-beta bot.

- `engine/` owns the rules: board state (bitboards), move generation, check
  detection, castling, en passant
- `watchdog/` owns the bot: search and evaluation
- the root package owns the GUI: rendering, input, turn handling

## Building

Requires Go and the [Ebiten system dependencies](https://ebitengine.org/en/documents/install.html)
(C compiler plus X11/GL development libraries on Linux).

```sh
make            # vet + build -> bin/watchdog-chess
./bin/watchdog-chess

# or just
go run .
```

Drag and drop pieces with the left mouse button. Which color you play is set
by `playAsWhite` in `game.go`. The
bot's search depth is `watchdogDepth` in `main.go`, this is 7 by default.

## Testing

The test suite is built around [perft](https://www.chessprogramming.org/Perft):
counting every legal-move path to a fixed depth and comparing against
[published reference values](https://www.chessprogramming.org/Perft_Results).
If the counts match, move generation is correct; if they don't, something
specific is broken. The positions live in `engine/perft_test.go` and cover
castling, pins, en passant, promotions, and checkmates.

```sh
make test-short   # shallow perft only, runs in ~0.1s
make test         # full suite, ~10s
make test-deep    # adds startpos depth 6 (119M nodes), ~2min
```

Equivalent raw commands:

```sh
go test -short ./...
go test ./...
PERFT_DEEP=1 go test -timeout 120m ./...
```

### Debugging a perft failure

A failing case automatically logs a per-root-move node breakdown
(`PerftDivide`). To locate the bug, compare that breakdown against a
known-good engine at the same position (e.g. stockfish: `position fen ...`
then `go perft N`), pick the first move whose count differs, play it, and
repeat one depth lower until the offending position is small enough to
inspect by hand. `engine/fen.go` provides `NewEngineFromFEN` for setting up
arbitrary positions.

## Benchmarks

```sh
make bench
```

Runs the move-generation and search-primitive benchmarks with allocation
stats. Record the numbers before and after any optimization work; `Perft(3)`
is the headline figure since it exercises the full generate/validate/make
pipeline.
