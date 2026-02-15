package logger

import (
	"io"
	"log/slog"
	"os"
)

// New creates a slog.Logger with the given level.
// If level is empty, uses INFO.
// If OPENCLAW_CURSOR_LOG_SILENT=true, returns a no-op logger that writes to io.Discard.
func New(level string) *slog.Logger {
	if os.Getenv("OPENCLAW_CURSOR_LOG_SILENT") == "true" {
		return slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}
	h := slog.NewTextHandler(os.Stderr, opts)
	return slog.New(h)
}
