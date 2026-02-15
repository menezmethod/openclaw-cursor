package streaming

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OpenAIDelta represents the delta in an OpenAI streaming chunk.
type OpenAIDelta struct {
	Content         string          `json:"content,omitempty"`
	ReasoningContent string         `json:"reasoning_content,omitempty"`
	ToolCalls       []OpenAIToolCall `json:"tool_calls,omitempty"`
}

// OpenAIToolCall is the OpenAI format for a tool call in streaming.
type OpenAIToolCall struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// OpenAIChunk is an OpenAI SSE chunk.
type OpenAIChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Delta        OpenAIDelta `json:"delta"`
		FinishReason interface{} `json:"finish_reason"`
	} `json:"choices"`
}

// Converter converts cursor-agent events to OpenAI SSE format.
type Converter struct {
	ID      string
	Created int64
	Model   string
	tracker DeltaTracker
}

// NewConverter creates a new SSE converter.
func NewConverter(model string) *Converter {
	return &Converter{
		ID:      fmt.Sprintf("openclaw-cursor-%d", 0),
		Created: 0,
		Model:   model,
	}
}

// ToSSEChunk converts a stream event to an OpenAI SSE chunk bytes.
// Returns nil if the event should not produce a chunk.
func (c *Converter) ToSSEChunk(event *StreamEvent) ([]byte, error) {
	if event == nil {
		return nil, nil
	}

	var delta OpenAIDelta

	if event.IsAssistantText() {
		text := event.ExtractText()
		d := c.tracker.NextText(text)
		if d == "" {
			return nil, nil
		}
		delta.Content = d
	}

	if event.IsThinking() {
		thinking := event.ExtractThinking()
		d := c.tracker.NextThinking(thinking)
		if d == "" {
			return nil, nil
		}
		delta.ReasoningContent = d
	}

	if event.IsToolCall() {
		tc := c.toolCallDelta(event)
		if tc != nil {
			delta.ToolCalls = []OpenAIToolCall{*tc}
		}
	}

	if delta.Content == "" && delta.ReasoningContent == "" && len(delta.ToolCalls) == 0 {
		return nil, nil
	}

	chunk := OpenAIChunk{
		ID:      c.ID,
		Object:  "chat.completion.chunk",
		Created: c.Created,
		Model:   c.Model,
		Choices: []struct {
			Index        int         `json:"index"`
			Delta        OpenAIDelta `json:"delta"`
			FinishReason interface{} `json:"finish_reason"`
		}{
			{Index: 0, Delta: delta, FinishReason: nil},
		},
	}
	b, err := json.Marshal(chunk)
	if err != nil {
		return nil, err
	}
	return []byte("data: " + string(b) + "\n\n"), nil
}

func (c *Converter) toolCallDelta(event *StreamEvent) *OpenAIToolCall {
	callID := event.CallID
	if callID == "" {
		callID = "unknown"
	}
	name := inferToolName(event)
	if name == "" {
		name = "tool"
	}
	args := "{}"
	if event.ToolCall != nil {
		// cursor-agent format: { "runCommandToolCall": { "args": {...} } }
		for key, v := range *event.ToolCall {
			if key == "args" || key == "result" {
				continue
			}
			if len(v) > 0 {
				var payload struct {
					Args map[string]interface{} `json:"args"`
				}
				if json.Unmarshal(v, &payload) == nil && payload.Args != nil {
					if b, err := json.Marshal(payload.Args); err == nil {
						args = string(b)
					}
				} else {
					args = string(v)
				}
				break
			}
		}
	}

	return &OpenAIToolCall{
		Index: 0,
		ID:    callID,
		Type:  "function",
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{Name: name, Arguments: args},
	}
}

func inferToolName(event *StreamEvent) string {
	if event.ToolCall == nil {
		return ""
	}
	for key := range *event.ToolCall {
		if key == "args" || key == "result" {
			continue
		}
		if strings.HasSuffix(key, "ToolCall") {
			base := strings.TrimSuffix(key, "ToolCall")
			if len(base) > 0 {
				return strings.ToLower(string(base[0])) + base[1:]
			}
		}
		return strings.ToLower(key)
	}
	return ""
}

// Done returns the SSE [DONE] chunk.
func (c *Converter) Done() []byte {
	return []byte("data: [DONE]\n\n")
}
