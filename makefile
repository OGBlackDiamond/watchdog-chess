.DEFAULT_GOAL := build

BINARY := watchdog-chess

fmt:
	go fmt ./...

lint: fmt
	golint ./...

vet: fmt
	go vet ./...

build: vet
	go build -o bin/$(BINARY) .

run:
	go run .

# quick correctness check (skips long perft cases)
test-short:
	go test -short ./...

# full package test suite
test-full:
	go test ./...

test: test-short

# short perft suite (skips deepest cases)
perft:
	go test ./internal/perft/ -short -v

# full perft suite, including deeper standard positions
perft-full:
	go test ./internal/perft/ -v

# compatibility alias for deeper perft runs
test-deep:
	go test ./internal/perft/ -v

# Run all Go benchmarks with allocation reporting. Packages without Benchmark...
# functions report no benchmarks.
bench:
	go test ./... -run '^$$' -bench . -benchmem

bench-perft:
	go test ./internal/perft/ -run '^$$' -bench . -benchmem

clean:
	rm -rf bin
