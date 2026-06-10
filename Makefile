VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo unknown)

run:
	go run ./cmd/gos

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -ldflags "-X github.com/cimoing/gos/internal/command.Version=$(VERSION) -X github.com/cimoing/gos/internal/command.Commit=$(COMMIT) -X github.com/cimoing/gos/internal/command.BuildDate=$(BUILD_DATE)" -o bin/gos ./cmd/gos
