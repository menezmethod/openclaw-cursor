# ---------------------------------------------------------------------------
# openclaw-cursor: multi-stage build for a portable Docker image
# Works on Linux (amd64/arm64), macOS (Apple Silicon + Intel), Windows (WSL2)
# ---------------------------------------------------------------------------

# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /bin/openclaw-cursor ./cmd/openclaw-cursor

# Stage 2: Install cursor-agent and OpenClaw into a Node image
FROM node:22-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates jq tini procps \
    && rm -rf /var/lib/apt/lists/*

# Install cursor-agent
RUN curl -fsSL https://cursor.com/install | bash

# Install OpenClaw globally
RUN npm install -g openclaw@latest 2>/dev/null || true

# Copy the proxy binary from the builder stage
COPY --from=builder /bin/openclaw-cursor /usr/local/bin/openclaw-cursor

# Copy scripts and skills
COPY scripts/ /opt/openclaw-cursor/scripts/
COPY skills/ /opt/openclaw-cursor/skills/

# Copy the entrypoint
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Default config directory (mount or provide via env)
RUN mkdir -p /root/.openclaw/logs /root/.openclaw/skills \
    && cp -r /opt/openclaw-cursor/skills/cursor-proxy /root/.openclaw/skills/ 2>/dev/null || true

# Proxy listens on 32125, OpenClaw gateway on 18789
EXPOSE 32125 18789

# Health check against the proxy
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -sf http://127.0.0.1:32125/health | jq -e '.status == "healthy"' > /dev/null

ENV OPENCLAW_CURSOR_PORT=32125 \
    OPENCLAW_CURSOR_LOG_LEVEL=info \
    OPENCLAW_CURSOR_TOOL_MODE=openclaw \
    OPENCLAW_CURSOR_ENABLE_THINKING=true \
    OPENCLAW_CURSOR_TIMEOUT_MS=300000

ENTRYPOINT ["tini", "--"]
CMD ["/entrypoint.sh"]
