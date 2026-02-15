#!/bin/bash
# Generates OpenClaw provider config for cursor models.
# Run: ./scripts/generate-openclaw-config.sh

PORT="${OPENCLAW_CURSOR_PORT:-32125}"
BASE_URL="http://127.0.0.1:${PORT}/v1"

cat << EOF
Add this to your ~/.openclaw/openclaw.json (under "models" -> "providers"):

  "cursor": {
    "baseUrl": "${BASE_URL}",
    "api": "openai-completions",
    "models": [
      { "id": "auto", "name": "Cursor Auto", "contextWindow": 200000, "maxTokens": 8192 },
      { "id": "composer-1.5", "name": "Composer 1.5", "contextWindow": 200000, "maxTokens": 8192 },
      { "id": "composer-1", "name": "Composer 1", "contextWindow": 200000, "maxTokens": 8192 },
      { "id": "gpt-5.3-codex", "name": "GPT-5.3 Codex", "contextWindow": 200000, "maxTokens": 8192 },
      { "id": "gpt-5.3-codex-high", "name": "GPT-5.3 Codex High", "contextWindow": 200000, "maxTokens": 8192 },
      { "id": "gpt-5.2", "name": "GPT-5.2", "contextWindow": 200000, "maxTokens": 8192 },
      { "id": "gpt-5.2-codex", "name": "GPT-5.2 Codex", "contextWindow": 200000, "maxTokens": 8192 },
      { "id": "sonnet-4.5", "name": "Claude 4.5 Sonnet", "contextWindow": 200000, "maxTokens": 8192 },
      { "id": "sonnet-4.5-thinking", "name": "Claude 4.5 Sonnet (Thinking)", "reasoning": true, "contextWindow": 200000, "maxTokens": 8192 },
      { "id": "opus-4.6", "name": "Claude 4.6 Opus", "contextWindow": 200000, "maxTokens": 8192 },
      { "id": "opus-4.6-thinking", "name": "Claude 4.6 Opus (Thinking)", "reasoning": true, "contextWindow": 200000, "maxTokens": 8192 }
    ]
  }

And set your default model:

  "agents": {
    "defaults": {
      "model": { "primary": "cursor/auto" }
    }
  }

Run 'openclaw-cursor models' for the full list.
EOF
