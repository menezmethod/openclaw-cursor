#!/bin/bash
set -e

echo "Installing openclaw-cursor proxy..."

# Check for cursor-agent
if ! command -v cursor-agent &> /dev/null; then
    echo "Installing cursor-agent..."
    curl -fsSL https://cursor.com/install | bash
fi

# Check for Go
if ! command -v go &> /dev/null; then
    echo "Go is required. Install from https://go.dev/dl/"
    exit 1
fi

# Build
cd "$(dirname "$0")/.."
go build -o bin/openclaw-cursor ./cmd/openclaw-cursor

# Create symlink
mkdir -p ~/.openclaw
mkdir -p ~/.local/bin 2>/dev/null || true
if [ -d ~/.local/bin ]; then
    ln -sf "$(pwd)/bin/openclaw-cursor" ~/.local/bin/openclaw-cursor
    echo "Installed to ~/.local/bin/openclaw-cursor"
else
    echo "Built to $(pwd)/bin/openclaw-cursor"
    echo "Add to PATH or: sudo ln -sf $(pwd)/bin/openclaw-cursor /usr/local/bin/openclaw-cursor"
fi

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "  1. Authenticate: openclaw-cursor login"
echo "  2. Start proxy: openclaw-cursor start"
echo "  3. Configure OpenClaw to use: http://127.0.0.1:32125"
echo ""
echo "If OpenClaw cron jobs fail with WhatsApp delivery errors, run (from this repo):"
echo "  node scripts/fix-cron-delivery.js --dry-run   # preview"
echo "  node scripts/fix-cron-delivery.js            # apply"
