# OpenClaw-Cursor Proxy

Use your **Cursor Pro subscription** as the AI backend for [OpenClaw](https://github.com/openclaw/openclaw). One proxy, all Cursor models, zero API keys.

```
You  -->  OpenClaw UI  -->  proxy (:32125)  -->  cursor-agent  -->  Cursor API
```

---

## Install (pick one)

### A) Docker (any OS)

```bash
git clone https://github.com/GabrieleRisso/openclaw-cursor.git
cd openclaw-cursor
docker compose up -d
```

Done. Proxy at `http://127.0.0.1:32125`, health at `/health`.

> **Windows?** Use Docker Desktop with WSL2. Same command in PowerShell or WSL terminal.

### B) From source (Linux / macOS)

```bash
# 1. Install cursor-agent
curl -fsSL https://cursor.com/install | bash

# 2. Install OpenClaw
curl -fsSL https://openclaw.ai/install.sh | bash

# 3. Clone and build the proxy
git clone https://github.com/GabrieleRisso/openclaw-cursor.git
cd openclaw-cursor
make build
make install-local

# 4. Log in to Cursor
openclaw-cursor login

# 5. Start the proxy
openclaw-cursor start
```

---

## Verify it works

```bash
# Health check
curl http://127.0.0.1:32125/health

# List models
curl http://127.0.0.1:32125/v1/models

# Send a chat request
curl http://127.0.0.1:32125/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"cursor/auto","messages":[{"role":"user","content":"Say hello"}]}'
```

---

## Connect OpenClaw

The proxy auto-generates the config on first run (Docker). For source installs, create `~/.openclaw/openclaw.json`:

```json
{
  "gateway": { "mode": "local", "port": 18789 },
  "models": {
    "mode": "merge",
    "providers": {
      "cursor": {
        "baseUrl": "http://127.0.0.1:32125/v1",
        "api": "openai-completions",
        "models": [
          { "id": "auto", "name": "Cursor Auto", "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "opus-4.6", "name": "Claude 4.6 Opus", "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "opus-4.6-thinking", "name": "Claude 4.6 Opus (Thinking)", "reasoning": true, "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "gpt-5.3-codex", "name": "GPT-5.3 Codex", "contextWindow": 200000, "maxTokens": 8192 }
        ]
      }
    }
  },
  "agents": { "defaults": { "model": { "primary": "cursor/auto" } } }
}
```

Then add the auth profile placeholder (`~/.openclaw/agents/main/agent/auth-profiles.json`):

```json
{
  "profiles": {
    "cursor:default": { "type": "api_key", "provider": "cursor", "key": "placeholder" }
  }
}
```

Start the gateway and open the dashboard:

```bash
openclaw gateway --port 18789
openclaw dashboard
```

---

## Docker details

### Authentication

The container needs your Cursor credentials. Pick one method:

| Method | How |
|--------|-----|
| **Mount auth dir** (default) | Auto-mounts `~/.config/cursor/` read-only |
| **Env vars** | `CURSOR_ACCESS_TOKEN` + `CURSOR_REFRESH_TOKEN` |
| **API key** | `CURSOR_API_KEY` |

```bash
# Example: pass tokens directly
CURSOR_ACCESS_TOKEN=eyJ... CURSOR_REFRESH_TOKEN=eyJ... docker compose up -d
```

### Modes

```bash
docker compose up -d                         # proxy only (default)
docker compose --profile gateway up -d       # proxy + OpenClaw UI
docker compose run --rm proxy test           # self-test
```

### Quick-start scripts

```bash
./docker/docker-run.sh           # Linux / macOS
.\docker\docker-run.ps1          # Windows PowerShell
```

### Auth file locations by OS

| OS | Path |
|----|------|
| Linux | `~/.config/cursor/auth.json` |
| macOS | `~/.cursor/auth.json` |
| Windows | `%APPDATA%\Cursor\auth.json` |

---

## CLI reference

```
openclaw-cursor login       # authenticate with Cursor
openclaw-cursor start       # start the proxy
openclaw-cursor start -d    # start in background
openclaw-cursor stop        # stop background proxy
openclaw-cursor status      # show auth + proxy status
openclaw-cursor models      # list all available models
openclaw-cursor test        # send a test request
openclaw-cursor version     # print version
```

---

## Models

Run `openclaw-cursor models` for the full list. Highlights:

| Model | ID |
|-------|----|
| Cursor Auto | `auto` |
| Claude 4.6 Opus | `opus-4.6` |
| Claude 4.6 Opus (Thinking) | `opus-4.6-thinking` |
| Claude 4.5 Sonnet | `sonnet-4.5` |
| GPT-5.3 Codex | `gpt-5.3-codex` |
| GPT-5.2 | `gpt-5.2` |
| Gemini 3 Pro | `gemini-3-pro` |

Use as `cursor/<id>` in OpenClaw (e.g. `/model cursor/opus-4.6`).

---

## Configuration

Config file: `~/.openclaw/cursor-proxy.json`

```json
{
  "port": 32125,
  "log_level": "info",
  "default_model": "auto",
  "enable_thinking": true,
  "timeout_ms": 300000
}
```

All settings can be overridden with `OPENCLAW_CURSOR_*` env vars.

---

## Troubleshooting

| Problem | Fix |
|---------|-----|
| cursor-agent not found | `curl -fsSL https://cursor.com/install \| bash` |
| Authentication failed | `openclaw-cursor login` |
| Quota exceeded | Check cursor.com/settings |
| Proxy not reachable | `openclaw-cursor start` |
| "No API key for cursor" | Add the auth-profiles.json placeholder above |

Debug mode: `OPENCLAW_CURSOR_LOG_LEVEL=debug openclaw-cursor start`

---

## API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/chat/completions` | POST | OpenAI-compatible chat |
| `/v1/models` | GET | List models |
| `/health` | GET | Health check |

## License

MIT
