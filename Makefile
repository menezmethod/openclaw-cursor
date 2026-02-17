VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)
HOME ?= $(shell echo $$HOME)
BINARY := $(HOME)/.local/bin/openclaw-cursor
PLIST := $(HOME)/Library/LaunchAgents/ai.openclaw.cursor-proxy.plist

.PHONY: build install install-local launchd-setup launchd-reload refresh cross test lint clean \
       docker docker-up docker-down docker-test docker-logs

build:
	go build -ldflags "$(LDFLAGS)" -o bin/openclaw-cursor ./cmd/openclaw-cursor

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/openclaw-cursor

install-local: build
	@mkdir -p $(HOME)/.local/bin
	@ln -sf "$(shell pwd)/bin/openclaw-cursor" $(BINARY)
	@echo "Installed to $(BINARY)"

launchd-setup: install-local
	@mkdir -p $(HOME)/.openclaw/logs
	@printf '%s\n' \
		'<?xml version="1.0" encoding="UTF-8"?>' \
		'<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' \
		'<plist version="1.0"><dict>' \
		'  <key>Label</key><string>ai.openclaw.cursor-proxy</string>' \
		'  <key>RunAtLoad</key><true/>' \
		'  <key>KeepAlive</key><true/>' \
		'  <key>ProgramArguments</key><array>' \
		'    <string>$(BINARY)</string>' \
		'    <string>start</string>' \
		'  </array>' \
		'  <key>StandardOutPath</key><string>$(HOME)/.openclaw/logs/cursor-proxy.log</string>' \
		'  <key>StandardErrorPath</key><string>$(HOME)/.openclaw/logs/cursor-proxy.err.log</string>' \
		'</dict></plist>' > $(PLIST)
	@echo "Created $(PLIST)"

launchd-reload:
	@launchctl unload $(PLIST) 2>/dev/null || true
	@launchctl load $(PLIST)
	@echo "Reloaded cursor-proxy (check $(HOME)/.openclaw/logs/cursor-proxy.err.log)"

refresh: launchd-setup launchd-reload
	@echo "Done. Cursor-proxy running via launchd from $(BINARY)"

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

# ── Docker targets ──────────────────────────────────────────────
docker:
	docker build -t openclaw-cursor:latest --build-arg VERSION=$(VERSION) .

docker-up: docker
	docker compose up -d proxy

docker-down:
	docker compose down

docker-test: docker
	docker compose run --rm proxy test

docker-logs:
	docker compose logs -f proxy
