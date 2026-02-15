package errors

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

// ParsedError represents a parsed cursor-agent error.
type ParsedError struct {
	Type        string `json:"error"`
	Message     string `json:"message"`
	Recoverable bool   `json:"recoverable"`
	Suggestion  string `json:"suggestion,omitempty"`
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// Parse parses cursor-agent stderr into a structured error.
func Parse(stderr string) *ParsedError {
	s := strings.ToLower(stripANSI(stderr))

	if strings.Contains(s, "usage limit") || strings.Contains(s, "quota") || strings.Contains(s, "exceeded") {
		return &ParsedError{
			Type:        "quota_exceeded",
			Message:     "Cursor quota exceeded. Check cursor.com/settings",
			Recoverable: false,
			Suggestion:  "Check your Cursor subscription and usage at cursor.com/settings",
		}
	}
	if strings.Contains(s, "not logged in") || strings.Contains(s, "auth") || strings.Contains(s, "unauthorized") || strings.Contains(s, "authentication failed") {
		return &ParsedError{
			Type:        "auth_failed",
			Message:     "Cursor authentication invalid",
			Recoverable: false,
			Suggestion:  "Run: openclaw-cursor login",
		}
	}
	if strings.Contains(s, "model not found") || strings.Contains(s, "invalid model") || strings.Contains(s, "cannot use this model") {
		return &ParsedError{
			Type:        "model_unavailable",
			Message:     "Model not available in your Cursor plan",
			Recoverable: false,
		}
	}
	if strings.Contains(s, "econnrefused") || strings.Contains(s, "connection refused") || strings.Contains(s, "network") || strings.Contains(s, "fetch failed") {
		return &ParsedError{
			Type:        "network_error",
			Message:     "Network error connecting to Cursor API",
			Recoverable: true,
			Suggestion:  "Check your internet connection and try again",
		}
	}

	return &ParsedError{
		Type:        "unknown",
		Message:     strings.TrimSpace(stderr),
		Recoverable: false,
	}
}

// OpenAIErrorResponse is the OpenAI API error format.
type OpenAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`
}

// ToOpenAIError formats a ParsedError as OpenAI API error JSON.
func ToOpenAIError(pe *ParsedError) OpenAIErrorResponse {
	msg := pe.Message
	if pe.Suggestion != "" {
		msg += ". " + pe.Suggestion
	}
	return OpenAIErrorResponse{
		Error: struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code,omitempty"`
		}{
			Message: msg,
			Type:    pe.Type,
			Code:    pe.Type,
		},
	}
}

// ToOpenAIErrorJSON returns the JSON bytes for the error response.
func ToOpenAIErrorJSON(pe *ParsedError) ([]byte, error) {
	return json.Marshal(ToOpenAIError(pe))
}

// Retry executes fn with exponential backoff for recoverable errors.
func Retry(ctx context.Context, maxAttempts int, fn func() error) error {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	var lastErr error
	backoff := time.Second
	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		pe := Parse(err.Error())
		if !pe.Recoverable || attempt == maxAttempts-1 {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}
	return lastErr
}
