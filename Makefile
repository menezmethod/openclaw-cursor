VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install cross test lint clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/openclaw-cursor ./cmd/openclaw-cursor

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/openclaw-cursor

cross:
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/openclaw-cursor-linux-amd64 ./cmd/openclaw-cursor
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/openclaw-cursor-linux-arm64 ./cmd/openclaw-cursor
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/openclaw-cursor-darwin-arm64 ./cmd/openclaw-cursor
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/openclaw-cursor-darwin-amd64 ./cmd/openclaw-cursor

test:
	go test -v -race ./...

lint:
	@which golangci-lint >/dev/null 2>&1 || (echo "Install golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

clean:
	rm -rf bin dist
