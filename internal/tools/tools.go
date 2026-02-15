package tools

import (
	"encoding/json"
	"strings"

	"github.com/menezmethod/openclaw-cursor/internal/streaming"
	"github.com/menezmethod/openclaw-cursor/internal/translator"
)

// AliasMap normalizes cursor-agent tool names to OpenClaw tool names.
// OpenClaw: exec, bash, shell, process (group:runtime) | read, write, edit, apply_patch (group:fs)
// cursor-agent: runCommand, bash, read, write, edit
var AliasMap = map[string]string{
	"runcommand":  "bash",
	"run_command": "bash",
	"runcommandtoolcall": "bash",
	"run_command_tool_call": "bash",
	"shell":       "bash",
	"shelltoolcall": "bash",
	"exec":        "bash",
	"bash":        "bash",
	"write":       "write",
	"edit":        "edit",
	"read":        "read",
	"apply_patch": "edit",
}

// OpenAIToolCall is the OpenAI format for a tool call.
type OpenAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// InterceptToolCall extracts tool call from cursor-agent event and converts to OpenAI format.
func InterceptToolCall(event *streaming.StreamEvent) (*OpenAIToolCall, error) {
	if event == nil || !event.IsToolCall() {
		return nil, nil
	}
	name := inferToolName(event)
	name = NormalizeName(name)
	callID := event.CallID
	if callID == "" {
		callID = "unknown"
	}
	args := "{}"
	if event.ToolCall != nil {
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
		ID:   callID,
		Type: "function",
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{Name: name, Arguments: args},
	}, nil
}

func inferToolName(event *streaming.StreamEvent) string {
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

// NormalizeName applies alias map for tool name.
func NormalizeName(name string) string {
	if name == "" {
		return "tool"
	}
	lower := strings.ToLower(name)
	if alias, ok := AliasMap[lower]; ok {
		return alias
	}
	return name
}

// FormatToolResult converts a tool message to TOOL_RESULT block for cursor-agent prompt.
func FormatToolResult(msg translator.Message) string {
	callID := msg.ToolCallID
	if callID == "" {
		callID = "unknown"
	}
	content := string(msg.Content)
	if len(content) >= 2 && content[0] == '"' {
		var s string
		if json.Unmarshal(msg.Content, &s) == nil {
			content = s
		}
	}
	return "TOOL_RESULT (call_id: " + callID + "): " + content
}

// LoopGuard prevents infinite tool loops.
type LoopGuard struct {
	fingerprints map[string]int
	maxRepeats   int
}

// NewLoopGuard creates a loop guard.
func NewLoopGuard(maxRepeats int) *LoopGuard {
	if maxRepeats <= 0 {
		maxRepeats = 3
	}
	return &LoopGuard{
		fingerprints: make(map[string]int),
		maxRepeats:   maxRepeats,
	}
}

// Fingerprint creates a fingerprint for a tool call.
func Fingerprint(name, args string) string {
	return name + ":" + args
}

// Record records a tool call and returns true if it should be allowed.
func (g *LoopGuard) Record(name, args string) bool {
	fp := Fingerprint(name, args)
	g.fingerprints[fp]++
	return g.fingerprints[fp] <= g.maxRepeats
}

// Reset clears the guard state.
func (g *LoopGuard) Reset() {
	g.fingerprints = make(map[string]int)
}
