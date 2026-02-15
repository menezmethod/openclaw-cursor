package streaming

import "encoding/json"

// StreamEvent represents a cursor-agent NDJSON stream event.
// Uses json.RawMessage for flexible parsing of nested structures.
type StreamEvent struct {
	Type    string          `json:"type"`
	Subtype string          `json:"subtype,omitempty"`
	Text    string          `json:"text,omitempty"`
	Message *StreamMessage  `json:"message,omitempty"`
	ToolCall *StreamToolCall `json:"tool_call,omitempty"`
	CallID  string          `json:"call_id,omitempty"`
}

// StreamMessage is the message content in assistant events.
type StreamMessage struct {
	Role    string           `json:"role"`
	Content []StreamContent  `json:"content,omitempty"`
}

// StreamContent is text or thinking content.
type StreamContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

// StreamToolCall wraps the tool call payload (keys vary: RunCommandToolCall, etc).
type StreamToolCall map[string]json.RawMessage

// IsAssistantText returns true if event has assistant text content.
func (e *StreamEvent) IsAssistantText() bool {
	if e.Type != "assistant" || e.Message == nil {
		return false
	}
	for _, c := range e.Message.Content {
		if c.Type == "text" && c.Text != "" {
			return true
		}
	}
	return false
}

// IsThinking returns true for thinking events (type=thinking or assistant with thinking content).
func (e *StreamEvent) IsThinking() bool {
	if e.Type == "thinking" {
		return true
	}
	if e.Type == "assistant" && e.Message != nil {
		for _, c := range e.Message.Content {
			if c.Type == "thinking" && c.Thinking != "" {
				return true
			}
		}
	}
	return false
}

// IsToolCall returns true for tool_call events.
func (e *StreamEvent) IsToolCall() bool {
	return e.Type == "tool_call"
}

// IsResult returns true for result events.
func (e *StreamEvent) IsResult() bool {
	return e.Type == "result"
}

// ExtractText returns text from assistant event.
func (e *StreamEvent) ExtractText() string {
	if e.Message == nil {
		return ""
	}
	var s string
	for _, c := range e.Message.Content {
		if c.Type == "text" {
			s += c.Text
		}
	}
	return s
}

// ExtractThinking returns thinking text from thinking or assistant event.
func (e *StreamEvent) ExtractThinking() string {
	if e.Type == "thinking" {
		return e.Text
	}
	if e.Message != nil {
		var s string
		for _, c := range e.Message.Content {
			if c.Type == "thinking" {
				s += c.Thinking
			}
		}
		return s
	}
	return ""
}
