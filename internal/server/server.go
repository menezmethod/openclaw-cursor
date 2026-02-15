package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/menezmethod/openclaw-cursor/internal/agent"
	"github.com/menezmethod/openclaw-cursor/internal/auth"
	"github.com/menezmethod/openclaw-cursor/internal/config"
	"github.com/menezmethod/openclaw-cursor/internal/errors"
	"github.com/menezmethod/openclaw-cursor/internal/models"
	"github.com/menezmethod/openclaw-cursor/internal/streaming"
	"github.com/menezmethod/openclaw-cursor/internal/translator"
)

// Server is the HTTP proxy server.
type Server struct {
	cfg    *config.Config
	log    *slog.Logger
	mux    *http.ServeMux
	server *http.Server
}

// New creates a new server.
func New(cfg *config.Config, log *slog.Logger) *Server {
	s := &Server{cfg: cfg, log: log, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /v1/chat/completions", s.handleChatCompletions)
	s.mux.HandleFunc("GET /v1/models", s.handleListModels)
	s.mux.HandleFunc("GET /health", s.handleHealth)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	authStatus := auth.VerifyAuth()
	cursorAgent := "unavailable"
	if _, err := agent.FindBinary(); err == nil {
		cursorAgent = "available"
	}
	status := map[string]interface{}{
		"status":        "healthy",
		"cursor_agent":  cursorAgent,
		"authenticated": authStatus.Authenticated,
		"proxy_version": "1.0.0",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleListModels(w http.ResponseWriter, r *http.Request) {
	list := models.ListOpenAI()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, errors.Parse(err.Error()))
		return
	}

	var req translator.ChatCompletionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeError(w, &errors.ParsedError{Type: "invalid_request", Message: "Invalid JSON body"})
		return
	}

	modelID, err := models.Resolve(req.Model)
	if err != nil {
		s.writeError(w, &errors.ParsedError{Type: "model_unavailable", Message: err.Error()})
		return
	}

	prompt := translator.BuildPrompt(req)
	timeout := time.Duration(s.cfg.TimeoutMs) * time.Millisecond

	// Use Background for non-streaming: request context can be cancelled when client
	// closes the connection (e.g. some HTTP clients), which would kill cursor-agent.
	// For streaming we pass r.Context() in the handler so client disconnect stops the stream.
	stream := req.Stream != nil && *req.Stream
	spawnCtx := context.Background()
	if stream {
		spawnCtx = r.Context()
	}
	proc, err := agent.Spawn(spawnCtx, agent.Options{
		Model:     modelID,
		Prompt:    prompt,
		Workspace: ".",
		Timeout:   timeout,
	})
	if err != nil {
		s.writeError(w, errors.Parse(err.Error()))
		return
	}
	// Only kill on early return; once Wait() succeeds the process has exited
	defer func() { _ = proc.Kill() }()

	if stream {
		s.handleStreaming(w, r, proc, modelID)
	} else {
		s.handleNonStreaming(w, proc, modelID)
	}
}

func (s *Server) handleStreaming(w http.ResponseWriter, r *http.Request, proc *agent.Process, modelID string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	conv := streaming.NewConverter(modelID)
	sc := streaming.NewScanner(proc.Stdout())

	go io.Copy(io.Discard, proc.Stderr()) // Drain stderr

	for sc.Scan() {
		select {
		case <-r.Context().Done():
			return
		default:
		}
		event, err := sc.Event()
		if err != nil {
			s.log.Debug("parse event", "err", err)
			continue
		}
		if event == nil {
			continue
		}
		chunk, err := conv.ToSSEChunk(event)
		if err != nil {
			continue
		}
		if len(chunk) > 0 {
			w.Write(chunk)
			flusher.Flush()
		}
	}
	w.Write(conv.Done())
	flusher.Flush()
	_ = proc.Wait() // Reap process and release context
}

func (s *Server) handleNonStreaming(w http.ResponseWriter, proc *agent.Process, modelID string) {
	var stdout, stderr []byte
	var stdoutErr error
	done := make(chan struct{})
	go func() {
		stdout, stdoutErr = io.ReadAll(proc.Stdout())
		done <- struct{}{}
	}()
	go func() {
		stderr, _ = io.ReadAll(proc.Stderr())
		done <- struct{}{}
	}()
	<-done
	<-done
	if stdoutErr != nil {
		s.writeError(w, errors.Parse(stdoutErr.Error()))
		return
	}
	if err := proc.Wait(); err != nil {
		pe := errors.Parse(string(stderr))
		if pe.Type == "unknown" {
			pe.Message = err.Error()
		}
		s.writeError(w, pe)
		return
	}

	// Parse all events and assemble response
	var content, reasoning string
	sc := streaming.NewScanner(bytes.NewReader(stdout))
	for sc.Scan() {
		event, _ := sc.Event()
		if event == nil {
			continue
		}
		if event.IsAssistantText() {
			content += event.ExtractText()
		}
		if event.IsThinking() {
			reasoning += event.ExtractThinking()
		}
	}

	msg := map[string]interface{}{"role": "assistant", "content": content}
	if reasoning != "" {
		msg["reasoning_content"] = reasoning
	}
	resp := map[string]interface{}{
		"id":      "openclaw-cursor-1",
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   modelID,
		"choices": []map[string]interface{}{
			{"index": 0, "message": msg, "finish_reason": "stop"},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) writeError(w http.ResponseWriter, pe *errors.ParsedError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(errors.ToOpenAIError(pe))
}

// Start runs the server with graceful shutdown.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		s.log.Info("proxy listening", "addr", addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error("server error", "err", err)
		}
	}()

	<-ctx.Done()
	s.log.Info("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.server.Shutdown(shutdownCtx)
}
