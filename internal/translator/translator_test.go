package translator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPrompt_SimpleMessages(t *testing.T) {
	req := ChatCompletionRequest{
		Model: "cursor/auto",
		Messages: []Message{
			{Role: "user", Content: json.RawMessage(`"Hello"`)},
			{Role: "assistant", Content: json.RawMessage(`"Hi there!"`)},
			{Role: "user", Content: json.RawMessage(`"How are you?"`)},
		},
	}
	prompt := BuildPrompt(req)
	assert.Contains(t, prompt, "USER: Hello")
	assert.Contains(t, prompt, "ASSISTANT: Hi there!")
	assert.Contains(t, prompt, "USER: How are you?")
}

func TestBuildPrompt_SystemMessage(t *testing.T) {
	req := ChatCompletionRequest{
		Messages: []Message{
			{Role: "system", Content: json.RawMessage(`"You are helpful."`)},
			{Role: "user", Content: json.RawMessage(`"Hi"`)},
		},
	}
	prompt := BuildPrompt(req)
	assert.Contains(t, prompt, "SYSTEM: You are helpful.")
	assert.Contains(t, prompt, "USER: Hi")
}

func TestBuildPrompt_ToolResult(t *testing.T) {
	req := ChatCompletionRequest{
		Messages: []Message{
			{Role: "assistant", Content: json.RawMessage(`""`), ToolCalls: []ToolCall{
				{ID: "call_1", Function: ToolCallFn{Name: "bash", Arguments: `{"command":"ls"}`}},
			}},
			{Role: "tool", ToolCallID: "call_1", Content: json.RawMessage(`"file1.txt\nfile2.txt"`)},
		},
	}
	prompt := BuildPrompt(req)
	assert.Contains(t, prompt, "tool_call(id: call_1, name: bash, args:")
	assert.Contains(t, prompt, "TOOL_RESULT (call_id: call_1): file1.txt")
	assert.Contains(t, prompt, "The above tool calls have been executed")
}

func TestBuildPrompt_WithTools(t *testing.T) {
	req := ChatCompletionRequest{
		Messages: []Message{
			{Role: "user", Content: json.RawMessage(`"List files"`)},
		},
		Tools: []ToolDefinition{
			{
				Type: "function",
				Function: &ToolDefFn{
					Name:        "bash",
					Description: "Run bash command",
					Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string"}}}`),
				},
			},
		},
	}
	prompt := BuildPrompt(req)
	assert.Contains(t, prompt, "You have access to the following tools")
	assert.Contains(t, prompt, "- bash: Run bash command")
}

func TestExtractTextContent_String(t *testing.T) {
	content := json.RawMessage(`"hello world"`)
	assert.Equal(t, "hello world", extractTextContent(content))
}

func TestExtractTextContent_Array(t *testing.T) {
	content := json.RawMessage(`[{"type":"text","text":"part1"},{"type":"text","text":"part2"}]`)
	assert.Equal(t, "part1\npart2", extractTextContent(content))
}
