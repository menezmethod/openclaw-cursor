# OpenClaw-Cursor Proxy

HTTP proxy that enables [OpenClaw](https://github.com/openclaw/openclaw) to use Cursor Pro models via the `cursor-agent` CLI. A single static Go binary with no runtime dependencies.

## Architecture

```
OpenClaw --> POST /v1/chat/completions --> openclaw-cursor proxy (:32125)
                    |
                    v
            cursor-agent (spawned per request)
                    |
                    v
            Cursor API (HTTPS)
```

- Accepts OpenAI-compatible requests
- Spawns `cursor-agent --output-format stream-json` per request
- Streams NDJSON responses as OpenAI SSE
- Supports thinking blocks and tool calling (OpenClaw-owned loop)

## Prerequisites

- **Go 1.22+** (for building)
- **cursor-agent** - `curl -fsSL https://cursor.com/install | bash`

## Installation

### Build from source

```bash
git clone https://github.com/menezmethod/openclaw-cursor.git
cd openclaw-cursor
make build
# Binary at bin/openclaw-cursor
```

### Install script

```bash
./scripts/install.sh
```

### Cross-compile

```bash
make cross
# Outputs to dist/openclaw-cursor-{linux,darwin}-{amd64,arm64}
```

## Quick Start

```bash
# 1. Authenticate with Cursor (opens browser)
openclaw-cursor login

# 2. Start the proxy
openclaw-cursor start

# 3. Configure OpenClaw - add to ~/.openclaw/openclaw.json:
```

```json
{
  "models": {
    "mode": "merge",
    "providers": {
      "cursor": {
        "baseUrl": "http://127.0.0.1:32125/v1",
        "api": "openai-completions",
        "models": [
          { "id": "auto", "name": "Cursor Auto", "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "sonnet-4.5-thinking", "name": "Claude 4.5 Sonnet (Thinking)", "reasoning": true, "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "opus-4.6", "name": "Claude 4.6 Opus", "contextWindow": 200000, "maxTokens": 8192 },
          { "id": "opus-4.6-thinking", "name": "Claude 4.6 Opus (Thinking)", "reasoning": true, "contextWindow": 200000, "maxTokens": 8192 }
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
```

```bash
# 4. Test
openclaw-cursor test
```

## OpenClaw Self-Onboarding

Use the proxy so OpenClaw runs its own onboarding wizard with Cursor models. OpenClaw will use your Cursor subscription to guide you through setup.

**Prerequisites:** [OpenClaw](https://docs.openclaw.ai/) installed (`npm install -g openclaw@latest`).

```bash
# 1. Start the proxy
openclaw-cursor start

# 2. In another terminal, run the onboarding wizard
openclaw onboard --install-daemon
```

When the wizard reaches **Model/Auth**, choose **Custom Provider** and enter:

- **Base URL:** `http://127.0.0.1:32125/v1`
- **Compatibility:** `openai`
- **API key:** leave empty (the proxy uses Cursor's auth via cursor-agent)

**Important:** OpenClaw requires an auth profile for every provider. Add a placeholder entry to `~/.openclaw/agents/main/agent/auth-profiles.json` under `profiles`:

```json
"cursor:default": {
  "type": "api_key",
  "provider": "cursor",
  "key": "placeholder"
}
```

The proxy ignores this key—it uses Cursor's auth via cursor-agent.

Alternatively, pre-configure Cursor in `~/.openclaw/openclaw.json` as shown in [Quick Start](#quick-start) before running the wizard.

### Default model

```bash
openclaw models set cursor/auto
```

### Switch models mid-chat

In the Control UI or chat channels, type:

| Command | Model |
|---------|-------|
| `/model cursor/auto` | Cursor Auto |
| `/model cursor/opus-4.6` | Claude 4.6 Opus |
| `/model cursor/opus-4.6-thinking` | Opus with thinking |
| `/model cursor/sonnet-4.5-thinking` | Sonnet with thinking |
| `/model cursor/gpt-5.3-codex` | GPT-5.3 Codex |

Or run `openclaw agent --message "..."` after setting your default with `openclaw models set cursor/auto`.

## CLI Commands

| Command | Description |
|---------|-------------|
| `login` | Launch OAuth flow for Cursor |
| `logout` | Clear local credentials |
| `status` | Check auth and proxy status |
| `start` | Start proxy (foreground) |
| `start --daemon` | Start proxy in background |
| `stop` | Stop daemon |
| `models` | List available models |
| `test` | Send test request |
| `version` | Print version |

## Configuration

Config file: `~/.openclaw/cursor-proxy.json`

```json
{
  "port": 32125,
  "log_level": "info",
  "tool_mode": "openclaw",
  "timeout_ms": 300000,
  "retry_attempts": 3,
  "default_model": "auto",
  "enable_thinking": true,
  "max_tool_loop_iterations": 10
}
```

Environment variables (override config):

- `OPENCLAW_CURSOR_PORT` - Port (default 32125)
- `OPENCLAW_CURSOR_LOG_LEVEL` - debug, info, warn, error
- `OPENCLAW_CURSOR_LOG_SILENT` - true to suppress logs
- `OPENCLAW_CURSOR_TOOL_MODE` - openclaw or proxy-exec
- `OPENCLAW_CURSOR_TIMEOUT_MS` - Request timeout
- `OPENCLAW_CURSOR_ENABLE_THINKING` - Enable thinking blocks

## Models

Run `openclaw-cursor models` for the full list. Key models:

- **Composer**: auto, composer-1.5, composer-1
- **GPT-5.3 Codex**: gpt-5.3-codex, gpt-5.3-codex-high, -low, -xhigh, -fast variants
- **GPT-5.2**: gpt-5.2, gpt-5.2-codex, gpt-5.2-high
- **Claude**: opus-4.6, opus-4.6-thinking, sonnet-4.5, sonnet-4.5-thinking — [Claude models overview](https://platform.claude.com/docs/en/about-claude/models/overview)
- **Other**: gemini-3-pro, gemini-3-flash, grok

## Troubleshooting

**"cursor-agent not found"**  
Install: `curl -fsSL https://cursor.com/install | bash`

**"Authentication failed"**  
Run `openclaw-cursor login` and complete the flow in your browser.

**"Quota exceeded"**  
Check your Cursor subscription at cursor.com/settings.

**"Proxy not reachable"**  
Ensure the proxy is running: `openclaw-cursor start`

**"No API key found for provider cursor" / chat hangs**  
OpenClaw requires an auth profile. Add `"cursor:default": {"type":"api_key","provider":"cursor","key":"placeholder"}` to `~/.openclaw/agents/main/agent/auth-profiles.json` under `profiles`. The proxy ignores the key.

**Debug logging**  
`OPENCLAW_CURSOR_LOG_LEVEL=debug openclaw-cursor start`

**OpenClaw cron: "Delivering to WhatsApp requires target..."**  
If isolated cron jobs fail with WhatsApp delivery errors, run this once **on your machine** (writes to `~/.openclaw`; automation environments often cannot):

```bash
cd ~/Development/openclaw-cursor
node scripts/fix-cron-delivery.js --dry-run   # preview
node scripts/fix-cron-delivery.js             # apply (creates backup first)
```

- Fixes: `"to": "User"` (no channel), Telegram without `to`, and `mode: "announce"` with no channel/to → all get `channel: "telegram"` and `to: "YOUR_TELEGRAM_ID"`. Override with `OPENCLAW_CRON_DEFAULT_TO` or edit the script.
- Backup: `~/.openclaw/cron/jobs.json.bak.<timestamp>`. Restore: `cp ~/.openclaw/cron/jobs.json.bak.<ts> ~/.openclaw/cron/jobs.json`.

**Set all cron jobs to cursor/auto (same as heartbeat)**  
To make every isolated cron job use `cursor/auto` (or another model), run **on your machine**:

```bash
cd ~/Development/openclaw-cursor
node scripts/set-cron-model.js --dry-run   # preview
node scripts/set-cron-model.js             # set all to cursor/auto
node scripts/set-cron-model.js cursor/opus-4.6   # or another model
```

## API Endpoints

- `POST /v1/chat/completions` - OpenAI-compatible chat (streaming and non-streaming)
- `GET /v1/models` - List models
- `GET /health` - Health check

## License

MIT
