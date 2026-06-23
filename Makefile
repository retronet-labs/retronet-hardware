# Makefile di RetroNet Hardware
# Richiede Go 1.25+ e, per lo sviluppo locale, un go.work che includa
# retronet-logic (vedi CONTRIBUTING.md).

.PHONY: all build test vet fmt fmt-check examples doc clean

all: fmt-check vet test

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || { echo "File non formattati:"; gofmt -l .; exit 1; }

examples:
	go run ./examples/...

doc:
	go doc ./...

clean:
	go clean ./...
