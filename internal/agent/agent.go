package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Options for spawning cursor-agent.
type Options struct {
	Model     string
	Prompt    string
	Workspace string
	Timeout   time.Duration
}

// Process wraps a cursor-agent subprocess.
type Process struct {
	cmd    *exec.Cmd
	stdout io.Reader
	stderr io.Reader
	cancel context.CancelFunc
}

// Stdout returns the process stdout reader.
func (p *Process) Stdout() io.Reader {
	return p.stdout
}

// Stderr returns the process stderr reader.
func (p *Process) Stderr() io.Reader {
	return p.stderr
}

// Wait waits for the process to exit.
func (p *Process) Wait() error {
	err := p.cmd.Wait()
	if p.cancel != nil {
		p.cancel()
	}
	return err
}

// Kill terminates the process.
func (p *Process) Kill() error {
	if p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}

// FindBinary locates the cursor-agent executable.
func FindBinary() (string, error) {
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

// Spawn starts cursor-agent with the given options.
// Context cancellation (e.g. client disconnect) will kill the subprocess.
func Spawn(ctx context.Context, opts Options) (*Process, error) {
	bin, err := FindBinary()
	if err != nil {
		return nil, err
	}

	workspace := opts.Workspace
	if workspace == "" {
		workspace, _ = os.Getwd()
	}
	if workspace == "" {
		workspace = "."
	}

	args := []string{
		"--print",
		"--output-format", "stream-json",
		"--stream-partial-output",
		"--trust", // Required for non-interactive use; avoids workspace trust prompt
		"--workspace", workspace,
		"--model", opts.Model,
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	// Don't call cancel here - it would kill the process. Process.Wait() calls cancel when done.

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = os.Environ()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start cursor-agent: %w", err)
	}

	// Write prompt to stdin and close
	go func() {
		stdin.Write([]byte(opts.Prompt))
		stdin.Close()
	}()

	return &Process{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		cancel: cancel,
	}, nil
}
