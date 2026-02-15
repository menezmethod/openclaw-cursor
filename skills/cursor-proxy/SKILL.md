---
name: cursor-proxy
description: Operate and troubleshoot the openclaw-cursor proxy (Cursor Pro models for OpenClaw)
metadata: {"openclaw":{"os":["darwin"]}}
---

# Cursor Proxy Operations

You help operate the openclaw-cursor proxy that lets OpenClaw use Cursor Pro models.

## Verify Latest Version

The proxy logs its version on boot. To check it:

1. **From terminal** (requires exec/bash):
   ```bash
   tail -5 ~/.openclaw/logs/cursor-proxy.err.log
   ```
   Look for `msg="proxy listening (version X)"` — X should match the latest git commit (e.g. `9a3447e-dirty`).

2. **From PATH**:
   ```bash
   openclaw-cursor version
   ```
   Shows the version of the binary at `~/.local/bin/openclaw-cursor` (or wherever `which openclaw-cursor` points).

3. **Health endpoint** (no exec needed):
   ```bash
   curl -s http://127.0.0.1:32125/health | jq .proxy_version
   ```

## Refresh (Build, Install, Reload)

To build the latest code, install to `~/.local/bin`, update the launchd plist, and reload:

```bash
cd ~/Development/openclaw-cursor
make refresh
```

This:
- Builds with version from git
- Symlinks `bin/openclaw-cursor` to `~/.local/bin/openclaw-cursor`
- Creates/updates `~/Library/LaunchAgents/ai.openclaw.cursor-proxy.plist`
- Reloads launchd (proxy restarts)

Run from the repo directory. Requires `make` and `go`.

## Paths

| What | Path |
|------|------|
| Binary (installed) | `~/.local/bin/openclaw-cursor` |
| Repo | `~/Development/openclaw-cursor` |
| Plist | `~/Library/LaunchAgents/ai.openclaw.cursor-proxy.plist` |
| Logs | `~/.openclaw/logs/cursor-proxy.log`, `cursor-proxy.err.log` |

## If User Reports Stale Proxy

If the user says the proxy isn't showing the latest version or changes:

1. Suggest: `cd ~/Development/openclaw-cursor && make refresh`
2. Or if they run manually: `make build` then copy `bin/openclaw-cursor` to `~/.local/bin/openclaw-cursor`
3. If using launchd: the plist must point at `~/.local/bin/openclaw-cursor` (or the actual binary path). Reload with `launchctl unload ~/Library/LaunchAgents/ai.openclaw.cursor-proxy.plist && launchctl load ...`

## Install This Skill

Copy to OpenClaw so the agent sees it:

```bash
mkdir -p ~/.openclaw/skills
cp -r ~/Development/openclaw-cursor/skills/cursor-proxy ~/.openclaw/skills/
```

Or install into workspace: `cp -r .../skills/cursor-proxy ~/.openclaw/workspace/skills/`
