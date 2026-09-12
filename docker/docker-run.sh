#!/bin/bash
# ---------------------------------------------------------------------------
# docker-run.sh — Quick-start script for Linux and macOS
#
# Detects your Cursor auth and launches the proxy container.
# Usage: ./docker/docker-run.sh [proxy|both|test]
# ---------------------------------------------------------------------------
set -euo pipefail

MODE="${1:-proxy}"
PORT="${OPENCLAW_CURSOR_PORT:-32125}"
IMAGE="openclaw-cursor:latest"
CONTAINER="openclaw-cursor-proxy"

# Detect Cursor auth directory
if [ "$(uname)" = "Darwin" ]; then
    AUTH_DIR="${HOME}/.cursor"
    AUTH_FILE="${AUTH_DIR}/auth.json"
    if [ ! -f "$AUTH_FILE" ]; then
        AUTH_DIR="${HOME}/Library/Application Support/Cursor"
    fi
else
    AUTH_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/cursor"
    AUTH_FILE="${AUTH_DIR}/auth.json"
fi

echo "Platform:  $(uname -s)/$(uname -m)"
echo "Auth dir:  $AUTH_DIR"
echo "Mode:      $MODE"
echo "Port:      $PORT"
echo ""

# Build if image doesn't exist
if ! docker image inspect "$IMAGE" > /dev/null 2>&1; then
    echo "Building Docker image..."
    SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
    docker build -t "$IMAGE" \
        --build-arg VERSION="$(git -C "$SCRIPT_DIR" describe --tags --always 2>/dev/null || echo dev)" \
        "$SCRIPT_DIR"
fi

# Stop existing container
docker rm -f "$CONTAINER" 2>/dev/null || true

# Auth volume mount
AUTH_MOUNT=""
if [ -d "$AUTH_DIR" ]; then
    AUTH_MOUNT="-v ${AUTH_DIR}:/root/.config/cursor:ro"
    echo "Mounting auth from: $AUTH_DIR"
else
    echo "WARNING: No Cursor auth directory found at $AUTH_DIR"
    echo "         Set CURSOR_ACCESS_TOKEN and CURSOR_REFRESH_TOKEN instead"
fi

# Run
GATEWAY_PORT="${OPENCLAW_GATEWAY_PORT:-18789}"
PORT_ARGS="-p ${PORT}:32125"
if [ "$MODE" = "both" ]; then
    PORT_ARGS="${PORT_ARGS} -p ${GATEWAY_PORT}:18789"
fi

# shellcheck disable=SC2086
docker run -d \
    --name "$CONTAINER" \
    ${PORT_ARGS} \
    ${AUTH_MOUNT} \
    -v openclaw-data:/root/.openclaw \
    -e CURSOR_ACCESS_TOKEN="${CURSOR_ACCESS_TOKEN:-}" \
    -e CURSOR_REFRESH_TOKEN="${CURSOR_REFRESH_TOKEN:-}" \
    -e CURSOR_API_KEY="${CURSOR_API_KEY:-}" \
    -e OPENCLAW_CURSOR_LOG_LEVEL="${OPENCLAW_CURSOR_LOG_LEVEL:-info}" \
    "$IMAGE" \
    "$MODE"

echo ""
echo "Container started: $CONTAINER"
echo "Health check: curl http://127.0.0.1:${PORT}/health"

# Wait for health
echo -n "Waiting for proxy..."
for i in $(seq 1 20); do
    if curl -sf "http://127.0.0.1:${PORT}/health" > /dev/null 2>&1; then
        echo " ready!"
        curl -s "http://127.0.0.1:${PORT}/health" | python3 -m json.tool 2>/dev/null || \
            curl -s "http://127.0.0.1:${PORT}/health"
        exit 0
    fi
    echo -n "."
    sleep 2
done
echo " timeout (check: docker logs $CONTAINER)"
exit 1
