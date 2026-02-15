package streaming

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeltaTracker(t *testing.T) {
	d := DeltaTracker{}
	assert.Equal(t, "hello", d.NextText("hello"))
	assert.Equal(t, " world", d.NextText("hello world"))
	assert.Equal(t, "", d.NextText("hello world"))
	assert.Equal(t, "!", d.NextText("hello world!"))
}

func TestConverter_ToSSEChunk(t *testing.T) {
	c := NewConverter("auto")
	e := &StreamEvent{
		Type: "assistant",
		Message: &StreamMessage{
			Role: "assistant",
			Content: []StreamContent{{Type: "text", Text: "Hi"}},
		},
	}
	chunk, err := c.ToSSEChunk(e)
	require.NoError(t, err)
	require.NotNil(t, chunk)
	assert.True(t, strings.HasPrefix(string(chunk), "data: "))
	assert.Contains(t, string(chunk), `"content":"Hi"`)
}

func TestConverter_Done(t *testing.T) {
	c := NewConverter("auto")
	done := c.Done()
	assert.Equal(t, "data: [DONE]\n\n", string(done))
}
