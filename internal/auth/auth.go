package auth

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

// AuthStatus represents authentication status.
type AuthStatus struct {
	Authenticated  bool   `json:"authenticated"`
	CredentialPath string `json:"credential_path,omitempty"`
	CursorAgent    string `json:"cursor_agent,omitempty"`
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// Possible credential file paths in priority order.
func authPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}
	return []string{
		filepath.Join(home, ".cursor", "cli-config.json"),
		filepath.Join(home, ".cursor", "auth.json"),
		filepath.Join(configHome, "cursor", "cli-config.json"),
	}
}

// VerifyAuth checks if Cursor credentials exist.
func VerifyAuth() AuthStatus {
	for _, p := range authPaths() {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return AuthStatus{Authenticated: true, CredentialPath: p}
		}
	}
	return AuthStatus{Authenticated: false}
}

// FindCursorAgent returns the path to the cursor-agent binary.
func FindCursorAgent() (string, error) {
	if p, err := exec.LookPath("cursor-agent"); err == nil {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err == nil {
		candidates := []string{
			filepath.Join(home, ".local", "bin", "cursor-agent"),
			"/usr/local/bin/cursor-agent",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		}
	}
	return "", fmt.Errorf("cursor-agent not found in PATH or common locations")
}

// StartLogin spawns cursor-agent login and polls for auth file creation.
func StartLogin(ctx context.Context) error {
	bin, err := FindCursorAgent()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, bin, "login")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run in background so we can poll for auth file while user completes login
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cursor-agent login: %w", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	deadline := time.Now().Add(5 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			cmd.Process.Kill()
			return ctx.Err()
		case err := <-done:
			if err != nil {
				return fmt.Errorf("cursor-agent login: %w", err)
			}
			// Process exited successfully; verify auth file exists
			if VerifyAuth().Authenticated {
				fmt.Fprintln(os.Stderr, "Login successful.")
				return nil
			}
			return fmt.Errorf("cursor-agent exited but auth file not found")
		case <-ticker.C:
			if time.Now().After(deadline) {
				cmd.Process.Kill()
				return fmt.Errorf("timeout waiting for auth file (5 minutes)")
			}
			if VerifyAuth().Authenticated {
				cmd.Process.Kill() // Process may still be running
				fmt.Fprintln(os.Stderr, "Login successful.")
				return nil
			}
		}
	}
}

// Logout removes local cursor-auth.json (if we stored one).
// The proxy delegates to cursor-agent; we don't actually remove
// ~/.cursor/* credentials (that would break cursor-agent).
// Per plan: Remove ~/.openclaw/cursor-auth.json if it exists.
func Logout() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	p := filepath.Join(home, ".openclaw", "cursor-auth.json")
	if _, err := os.Stat(p); err == nil {
		return os.Remove(p)
	}
	return nil
}

// GetStatus returns combined auth and cursor-agent status.
func GetStatus() AuthStatus {
	status := VerifyAuth()
	if bin, err := FindCursorAgent(); err == nil {
		status.CursorAgent = bin
	}
	return status
}
