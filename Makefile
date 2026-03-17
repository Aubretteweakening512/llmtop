.PHONY: build run test lint install release clean

VERSION ?= dev

build:
	go build -ldflags="-X main.version=$(VERSION)" -o llmtop ./cmd/llmtop

run:
	go run ./cmd/llmtop

test:
	go test -race -cover ./...

lint:
	golangci-lint run

install:
	go install -ldflags="-X main.version=$(VERSION)" ./cmd/llmtop

release:
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -f llmtop
	rm -rf dist/
