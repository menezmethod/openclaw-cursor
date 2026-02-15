package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Config holds proxy configuration.
type Config struct {
	Port                   int    `json:"port"`
	LogLevel               string `json:"log_level"`
	ToolMode               string `json:"tool_mode"`
	TimeoutMs              int    `json:"timeout_ms"`
	RetryAttempts          int    `json:"retry_attempts"`
	CursorAgentPath        string `json:"cursor_agent_path"`
	DefaultModel           string `json:"default_model"`
	EnableThinking         bool   `json:"enable_thinking"`
	MaxToolLoopIterations  int    `json:"max_tool_loop_iterations"`
}

// Default returns default configuration.
func Default() *Config {
	return &Config{
		Port:                  32125,
		LogLevel:              "info",
		ToolMode:              "openclaw",
		TimeoutMs:             300000,
		RetryAttempts:         3,
		CursorAgentPath:       "",
		DefaultModel:          "auto",
		EnableThinking:        true,
		MaxToolLoopIterations: 10,
	}
}

// ConfigFilePath returns the path to the config file.
func ConfigFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".openclaw", "cursor-proxy.json")
}

// Load reads config from file and merges with environment variables.
// Env vars override file config.
func Load() (*Config, error) {
	cfg := Default()

	path := ConfigFilePath()
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parse config %s: %w", path, err)
			}
		}
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("OPENCLAW_CURSOR_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Port = p
		}
	}
	if v := os.Getenv("OPENCLAW_CURSOR_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("OPENCLAW_CURSOR_TOOL_MODE"); v != "" {
		cfg.ToolMode = v
	}
	if v := os.Getenv("OPENCLAW_CURSOR_TIMEOUT_MS"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.TimeoutMs = p
		}
	}
	if v := os.Getenv("OPENCLAW_CURSOR_RETRY_ATTEMPTS"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.RetryAttempts = p
		}
	}
	if v := os.Getenv("OPENCLAW_CURSOR_CURSOR_AGENT_PATH"); v != "" {
		cfg.CursorAgentPath = v
	}
	if v := os.Getenv("OPENCLAW_CURSOR_DEFAULT_MODEL"); v != "" {
		cfg.DefaultModel = v
	}
	if v := os.Getenv("OPENCLAW_CURSOR_ENABLE_THINKING"); v != "" {
		cfg.EnableThinking = v == "true" || v == "1"
	}
	if v := os.Getenv("OPENCLAW_CURSOR_MAX_TOOL_LOOP_ITERATIONS"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.MaxToolLoopIterations = p
		}
	}
}
