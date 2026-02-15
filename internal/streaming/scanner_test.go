package streaming

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanner(t *testing.T) {
	input := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Hello"}]}}
{"type":"thinking","text":"let me think"}
`
	sc := NewScanner(strings.NewReader(input))
	require.True(t, sc.Scan())
	e, err := sc.Event()
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.Equal(t, "assistant", e.Type)
	assert.True(t, e.IsAssistantText())
	assert.Equal(t, "Hello", e.ExtractText())

	require.True(t, sc.Scan())
	e, err = sc.Event()
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.Equal(t, "thinking", e.Type)
	assert.True(t, e.IsThinking())
	assert.Equal(t, "let me think", e.ExtractThinking())

	assert.False(t, sc.Scan())
}

func TestScanner_EmptyLines(t *testing.T) {
	input := `{"type":"result","subtype":"success"}

`
	sc := NewScanner(strings.NewReader(input))
	require.True(t, sc.Scan())
	e, err := sc.Event()
	require.NoError(t, err)
	assert.Equal(t, "result", e.Type)
}
