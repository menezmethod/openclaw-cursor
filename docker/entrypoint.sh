#!/bin/bash
set -euo pipefail

# ---------------------------------------------------------------------------
# openclaw-cursor Docker entrypoint
# Handles auth injection, config generation, and service startup.
# ---------------------------------------------------------------------------

CONFIG_DIR="${HOME}/.openclaw"
AUTH_FILE="${XDG_CONFIG_HOME:-$HOME/.config}/cursor/auth.json"

echo "╔══════════════════════════════════════════════════════════╗"
echo "║       openclaw-cursor proxy — Docker container          ║"
echo "╚══════════════════════════════════════════════════════════╝"

# ── 1. Auth setup ──────────────────────────────────────────────
# Accept auth via: mounted file, env vars, or CURSOR_API_KEY
if [ -f "$AUTH_FILE" ]; then
    echo "[auth] Found auth.json at $AUTH_FILE"
elif [ -n "${CURSOR_ACCESS_TOKEN:-}" ] && [ -n "${CURSOR_REFRESH_TOKEN:-}" ]; then
    echo "[auth] Injecting auth from environment variables"
    mkdir -p "$(dirname "$AUTH_FILE")"
    cat > "$AUTH_FILE" <<AUTHEOF
{
  "accessToken": "${CURSOR_ACCESS_TOKEN}",
  "refreshToken": "${CURSOR_REFRESH_TOKEN}"
}
AUTHEOF
    chmod 600 "$AUTH_FILE"
elif [ -n "${CURSOR_API_KEY:-}" ]; then
    echo "[auth] Using CURSOR_API_KEY environment variable"
    mkdir -p "$(dirname "$AUTH_FILE")"
    cat > "$AUTH_FILE" <<AUTHEOF
{
  "apiKey": "${CURSOR_API_KEY}"
}
AUTHEOF
    chmod 600 "$AUTH_FILE"
else
    echo "[auth] WARNING: No Cursor credentials found."
    echo "       Provide one of:"
    echo "         - Mount ~/.config/cursor/auth.json"
    echo "         - Set CURSOR_ACCESS_TOKEN + CURSOR_REFRESH_TOKEN"
    echo "         - Set CURSOR_API_KEY"
fi

# ── 2. Proxy config ───────────────────────────────────────────
PROXY_CONFIG="${CONFIG_DIR}/cursor-proxy.json"
if [ ! -f "$PROXY_CONFIG" ]; then
    echo "[config] Generating proxy config"
    mkdir -p "$CONFIG_DIR"
    cat > "$PROXY_CONFIG" <<CFGEOF
{
  "port": ${OPENCLAW_CURSOR_PORT:-32125},
  "log_level": "${OPENCLAW_CURSOR_LOG_LEVEL:-info}",
  "tool_mode": "${OPENCLAW_CURSOR_TOOL_MODE:-openclaw}",
  "timeout_ms": ${OPENCLAW_CURSOR_TIMEOUT_MS:-300000},
  "retry_attempts": ${OPENCLAW_CURSOR_RETRY_ATTEMPTS:-3},
  "default_model": "${OPENCLAW_CURSOR_DEFAULT_MODEL:-auto}",
  "enable_thinking": ${OPENCLAW_CURSOR_ENABLE_THINKING:-true},
  "max_tool_loop_iterations": ${OPENCLAW_CURSOR_MAX_TOOL_LOOP_ITERATIONS:-10}
}
CFGEOF
fi

# ── 3. OpenClaw provider config ───────────────────────────────
OPENCLAW_CONFIG="${CONFIG_DIR}/openclaw.json"
if [ ! -f "$OPENCLAW_CONFIG" ]; then
    echo "[config] Generating OpenClaw provider config for Cursor"
    PORT="${OPENCLAW_CURSOR_PORT:-32125}"
    cat > "$OPENCLAW_CONFIG" <<OCONF
{
  "gateway": {
    "mode": "local",
    "port": ${OPENCLAW_GATEWAY_PORT:-18789}
  },
  "models": {
    "mode": "merge",
    "providers": {
      "cursor": {
        "baseUrl": "http://127.0.0.1:${PORT}/v1",
        "api": "openai-completions",
        "models": [
          { "id": "auto", "name": "Cursor Auto", "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "sonnet-4.5-thinking", "name": "Claude 4.5 Sonnet (Thinking)", "reasoning": true, "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "opus-4.6", "name": "Claude 4.6 Opus", "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "opus-4.6-thinking", "name": "Claude 4.6 Opus (Thinking)", "reasoning": true, "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "gpt-5.3-codex", "name": "GPT-5.3 Codex", "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "gpt-5.3-codex-high", "name": "GPT-5.3 Codex High", "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "gpt-5.2", "name": "GPT-5.2", "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "gpt-5.2-codex", "name": "GPT-5.2 Codex", "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "sonnet-4.5", "name": "Claude 4.5 Sonnet", "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "opus-4.5", "name": "Claude 4.5 Opus", "contextWindow": 200000, "maxTokens": 8192 }
        ]
      }
    }
  },
  "agents": {
    "defaults": {
      "model": { "primary": "cursor/auto" }
    }
  }
}
OCONF
fi

# ── 4. Auth profile placeholder ───────────────────────────────
AUTH_PROFILES="${CONFIG_DIR}/agents/main/agent/auth-profiles.json"
if [ ! -f "$AUTH_PROFILES" ]; then
    echo "[config] Creating auth profile placeholder"
    mkdir -p "$(dirname "$AUTH_PROFILES")"
    cat > "$AUTH_PROFILES" <<APEOF
{
  "profiles": {
    "cursor:default": {
      "type": "api_key",
      "provider": "cursor",
      "key": "placeholder"
    }
  }
}
APEOF
fi

# ── 5. Start ──────────────────────────────────────────────────
MODE="${1:-proxy}"

case "$MODE" in
    proxy)
        echo "[start] Starting openclaw-cursor proxy on port ${OPENCLAW_CURSOR_PORT:-32125}"
        exec openclaw-cursor start
        ;;
    both)
        echo "[start] Starting proxy + OpenClaw gateway"
        openclaw-cursor start &
        PROXY_PID=$!
        sleep 3
        if curl -sf http://127.0.0.1:${OPENCLAW_CURSOR_PORT:-32125}/health > /dev/null 2>&1; then
            echo "[start] Proxy healthy, starting OpenClaw gateway"
            exec openclaw gateway --port "${OPENCLAW_GATEWAY_PORT:-18789}"
        else
            echo "[error] Proxy failed to start"
            kill $PROXY_PID 2>/dev/null || true
            exit 1
        fi
        ;;
    test)
        echo "[test] Running proxy self-test"
        openclaw-cursor start &
        sleep 3
        openclaw-cursor test
        EXIT_CODE=$?
        echo "[test] Exit code: $EXIT_CODE"
        exit $EXIT_CODE
        ;;
    shell)
        exec /bin/bash
        ;;
    *)
        exec "$@"
        ;;
esac
