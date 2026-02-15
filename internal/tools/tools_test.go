package tools

import (
	"encoding/json"
	"testing"

	"github.com/menezmethod/openclaw-cursor/internal/streaming"
	"github.com/menezmethod/openclaw-cursor/internal/translator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeName(t *testing.T) {
	assert.Equal(t, "bash", NormalizeName("runcommand"))
	assert.Equal(t, "bash", NormalizeName("RunCommand"))
	assert.Equal(t, "write", NormalizeName("write"))
	assert.Equal(t, "tool", NormalizeName(""))
}

func TestFormatToolResult(t *testing.T) {
	msg := translator.Message{
		ToolCallID: "call_123",
		Content:    json.RawMessage(`"file1.txt"`),
	}
	s := FormatToolResult(msg)
	assert.Equal(t, "TOOL_RESULT (call_id: call_123): file1.txt", s)
}

func TestLoopGuard(t *testing.T) {
	g := NewLoopGuard(2)
	assert.True(t, g.Record("bash", `{"command":"ls"}`))
	assert.True(t, g.Record("bash", `{"command":"ls"}`))
	assert.False(t, g.Record("bash", `{"command":"ls"}`))

	g.Reset()
	assert.True(t, g.Record("bash", `{"command":"ls"}`))
}

func TestInterceptToolCall(t *testing.T) {
	raw := json.RawMessage(`{"command":"ls -la"}`)
	event := &streaming.StreamEvent{
		Type:   "tool_call",
		CallID: "call_1",
		ToolCall: &streaming.StreamToolCall{
			"runCommandToolCall": json.RawMessage(`{"args":{"command":"ls -la"}}`),
		},
	}
	tc, err := InterceptToolCall(event)
	require.NoError(t, err)
	require.NotNil(t, tc)
	assert.Equal(t, "call_1", tc.ID)
	assert.Equal(t, "bash", tc.Function.Name)
	assert.Contains(t, tc.Function.Arguments, "ls -la")
	_ = raw
}
