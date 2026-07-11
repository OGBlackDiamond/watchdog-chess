# watchdog-chess

A UCI chess engine written in Go.

## Running

Run the engine directly:

```sh
go run .
```

Or build a binary:

```sh
make build
./bin/watchdog-chess
```

The engine speaks UCI, so it expects commands such as `uci`, `isready`, `position`, `go`, and `quit` on stdin.

## Testing

Fast test pass:

```sh
make test
```

Full test pass:

```sh
make test-full
```

Move generation is validated with standard perft positions. The short perft suite skips the deepest cases:

```sh
make perft
```

Run the full perft suite, including deeper positions:

```sh
make perft-full
```

## Benchmarking

Run all Go benchmarks with allocation reporting:

```sh
make bench
```

Run benchmarks only for the perft package:

```sh
make bench-perft
```

If no `Benchmark...` functions exist in a package, Go will simply report that there are no benchmarks for that package.

## Useful Make Targets

- `make run` - run the engine with `go run .`
- `make build` - build `bin/watchdog-chess`
- `make test` - run `go test ./... -short`
- `make test-full` - run `go test ./...`
- `make perft` - run the short perft suite
- `make perft-full` - run the full perft suite
- `make bench` - run all benchmarks with `-benchmem`
- `make vet` - run `go vet ./...`
- `make fmt` - format Go files
- `make clean` - remove build artifacts
