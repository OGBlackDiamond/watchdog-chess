.DEFAULT_GOAL := build

fmt:
	go fmt ./...

lint: fmt
	golint ./...

vet: fmt
	go vet ./...

build: vet
	go build -o bin/watchdog-chess .

# quick correctness check (shallow perft only, a few seconds)
test-short:
	go test -short ./...

# full perft suite (excludes tierDeep cases)
test: vet
	go test ./...

# deep perft cases too; can take a long time on a slow move generator
test-deep:
	PERFT_DEEP=1 go test -timeout 120m ./...

# move generation speed baseline: record ns/op and allocs/op before and
# after any optimization
bench:
	go test ./engine -bench . -benchmem -run '^$$'

clean:
	rm -rf bin
