package translator

import (
	"encoding/json"
	"strings"
)

// ContentBlock represents a content part (OpenAI format).
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL any    `json:"image_url,omitempty"`
}

// Message represents an OpenAI chat message.
type Message struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content,omitempty"` // string or []ContentBlock
	ToolCalls  []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
}

// ToolCall represents an OpenAI tool call.
type ToolCall struct {
	ID       string     `json:"id"`
	Type     string     `json:"type"`
	Function ToolCallFn `json:"function"`
}

// ToolCallFn is the function part of a tool call.
type ToolCallFn struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolDefinition represents an OpenAI tool definition.
type ToolDefinition struct {
	Type     string       `json:"type"`
	Function *ToolDefFn   `json:"function,omitempty"`
}

// ToolDefFn is the function part of a tool definition.
type ToolDefFn struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// ChatCompletionRequest mirrors OpenAI request format.
type ChatCompletionRequest struct {
	Model       string          `json:"model"`
	Messages    []Message       `json:"messages"`
	Stream      *bool           `json:"stream,omitempty"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
	ToolChoice  any             `json:"tool_choice,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
}

func extractTextContent(content json.RawMessage) string {
	if len(content) == 0 {
		return ""
	}
	if content[0] == '"' {
		var s string
		if json.Unmarshal(content, &s) == nil {
			return s
		}
		return ""
	}
	if content[0] == '[' {
		var blocks []ContentBlock
		if json.Unmarshal(content, &blocks) != nil {
			return ""
		}
		var parts []string
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				parts = append(parts, b.Text)
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

// openClawToCursorTool maps OpenClaw tool names to cursor-agent equivalents.
// exec/shell -> bash; apply_patch -> edit; everything else passes through.
func openClawToCursorTool(name string) string {
	switch strings.ToLower(name) {
	case "exec", "shell":
		return "bash"
	case "apply_patch":
		return "edit"
	default:
		return name
	}
}

// BuildPrompt converts OpenAI chat messages to cursor-agent text format.
func BuildPrompt(req ChatCompletionRequest) string {
	var lines []string

	if len(req.Tools) > 0 {
		var toolDescs []string
		seen := make(map[string]bool)
		for _, t := range req.Tools {
			fn := t.Function
			if fn == nil {
				continue
			}
			name := fn.Name
			if name == "" {
				name = "unknown"
			}
			// Map OpenClaw tool names to cursor-agent equivalents (exec/shell → bash)
			cursorName := openClawToCursorTool(name)
			if seen[cursorName] {
				continue
			}
			seen[cursorName] = true
			desc := fn.Description
			if name != cursorName {
				desc = "[" + name + "→" + cursorName + "] " + desc
			}
			paramStr := "{}"
			if len(fn.Parameters) > 0 {
				paramStr = string(fn.Parameters)
			}
			toolDescs = append(toolDescs, "- "+cursorName+": "+desc+"\n  Parameters: "+paramStr)
		}
		if len(toolDescs) > 0 {
			lines = append(lines, "SYSTEM: You have access to the following tools. When you need to use one, respond with a tool_call in the standard OpenAI format.\n"+
				"Tool guidance (OpenClaw compatibility):\n"+
				"- exec, shell → use bash for running commands\n"+
				"- prefer write/edit for file changes; use bash for commands/tests\n"+
				"- For browser, cron, gateway, web_search, web_fetch, message, nodes, sessions_*: output the tool_call; OpenClaw executes these.\n\n"+
				"Available tools:\n"+
				strings.Join(toolDescs, "\n"))
		}
	}

	hasToolResults := false
	for _, msg := range req.Messages {
		role := msg.Role
		if role == "" {
			role = "user"
		}

		if role == "tool" {
			hasToolResults = true
			callID := msg.ToolCallID
			if callID == "" {
				callID = "unknown"
			}
			body := extractTextContent(msg.Content)
			if body == "" {
				body = string(msg.Content)
			}
			lines = append(lines, "TOOL_RESULT (call_id: "+callID+"): "+body)
			continue
		}

		if role == "assistant" && len(msg.ToolCalls) > 0 {
			var tcTexts []string
			for _, tc := range msg.ToolCalls {
				fn := tc.Function
				args := fn.Arguments
				if args == "" {
					args = "{}"
				}
				tcTexts = append(tcTexts, "tool_call(id: "+tc.ID+", name: "+fn.Name+", args: "+args+")")
			}
			text := extractTextContent(msg.Content)
			assistantLine := "ASSISTANT: "
			if text != "" {
				assistantLine += text + "\n"
			}
			assistantLine += strings.Join(tcTexts, "\n")
			lines = append(lines, assistantLine)
			continue
		}

		content := extractTextContent(msg.Content)
		if content != "" {
			lines = append(lines, strings.ToUpper(role)+": "+content)
		}
	}

	if hasToolResults {
		lines = append(lines, "The above tool calls have been executed. Continue your response based on these results.")
	}

	return strings.Join(lines, "\n\n")
}
