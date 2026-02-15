package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/menezmethod/openclaw-cursor/internal/auth"
	"github.com/menezmethod/openclaw-cursor/internal/config"
	"github.com/menezmethod/openclaw-cursor/internal/logger"
	"github.com/menezmethod/openclaw-cursor/internal/models"
	"github.com/menezmethod/openclaw-cursor/internal/server"
)

func runLogin() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return auth.StartLogin(ctx)
}

func runLogout() error {
	return auth.Logout()
}

func runStatus() error {
	status := auth.GetStatus()
	fmt.Println("Authentication:", status.Authenticated)
	if status.CredentialPath != "" {
		fmt.Println("Credentials:", status.CredentialPath)
	}
	if status.CursorAgent != "" {
		fmt.Println("cursor-agent:", status.CursorAgent)
	} else {
		fmt.Println("cursor-agent: not found")
	}

	cfg, _ := config.Load()
	if cfg == nil {
		cfg = config.Default()
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/health", cfg.Port)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Proxy: not running")
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		fmt.Println("Proxy: running on port", cfg.Port)
	}
	return nil
}

func runStart(daemon bool) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	log := logger.New(cfg.LogLevel)

	if daemon {
		return runDaemon(cfg)
	}

	srv := server.New(cfg, log, version)
	return srv.Start()
}

func runDaemon(cfg *config.Config) error {
	home, _ := os.UserHomeDir()
	openclawDir := filepath.Join(home, ".openclaw")
	os.MkdirAll(openclawDir, 0755)
	pidFile := filepath.Join(openclawDir, "cursor-proxy.pid")

	// Check if already running
	if data, err := os.ReadFile(pidFile); err == nil {
		pid, _ := strconv.Atoi(string(data))
		if pid > 0 {
			proc, _ := os.FindProcess(pid)
			if proc != nil {
				// Check if process exists
				if err := proc.Signal(os.Signal(nil)); err == nil {
					fmt.Fprintf(os.Stderr, "Proxy already running (PID %d)\n", pid)
					return nil
				}
			}
		}
	}

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable: %w", err)
	}

	cmd := exec.Command(self, "start")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.Dir = "/"
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}
	pid := cmd.Process.Pid
	os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
	fmt.Printf("Proxy started (PID %d), listening on port %d\n", pid, cfg.Port)
	return nil
}

func runStop() error {
	home, _ := os.UserHomeDir()
	pidFile := filepath.Join(home, ".openclaw", "cursor-proxy.pid")
	data, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Println("Proxy not running (no PID file)")
		return nil
	}
	pid, _ := strconv.Atoi(string(data))
	if pid <= 0 {
		return nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Println("Proxy not running")
		os.Remove(pidFile)
		return nil
	}
	if err := proc.Signal(os.Interrupt); err != nil {
		proc.Kill()
	}
	os.Remove(pidFile)
	fmt.Println("Proxy stopped")
	return nil
}

func runModels(jsonOut bool) error {
	list := models.ListOpenAI()
	if jsonOut {
		b, _ := json.MarshalIndent(list, "", "  ")
		fmt.Println(string(b))
		return nil
	}
	fmt.Printf("%-40s %s\n", "ID", "Name")
	fmt.Println("----------------------------------------")
	for _, m := range list.Data {
		model := models.Registry[m.ID]
		name := m.ID
		if model.Name != "" {
			name = model.Name
		}
		fmt.Printf("%-40s %s\n", m.ID, name)
	}
	return nil
}

func runTest() error {
	cfg, _ := config.Load()
	if cfg == nil {
		cfg = config.Default()
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/v1/chat/completions", cfg.Port)

	body := map[string]interface{}{
		"model": "cursor/auto",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Say 'test' in one word"},
		},
		"stream": false,
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("proxy not reachable: %w\nMake sure the proxy is running: openclaw-cursor start", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("request failed (status %d): %v", resp.StatusCode, errBody)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("invalid response: %w", err)
	}
	choices, _ := result["choices"].([]interface{})
	if len(choices) > 0 {
		msg := choices[0].(map[string]interface{})["message"]
		if m, ok := msg.(map[string]interface{}); ok {
			content := m["content"]
			fmt.Println("Response:", content)
		}
	}
	fmt.Println("Test passed.")
	return nil
}

func runVersion() error {
	fmt.Println(version)
	return nil
}
