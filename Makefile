run:
	go run ./cmd/gos

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o bin/gos ./cmd/gos
