package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"cursor/auto", "auto", false},
		{"cursor-acp/sonnet-4.5", "sonnet-4.5", false},
		{"auto", "auto", false},
		{"cursor/gpt-5.3-codex-high", "gpt-5.3-codex-high", false},
		{"cursor/unknown-model", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Resolve(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestListOpenAI(t *testing.T) {
	list := ListOpenAI()
	assert.Equal(t, "list", list.Object)
	assert.GreaterOrEqual(t, len(list.Data), 30)
	for _, m := range list.Data {
		assert.NotEmpty(t, m.ID)
		assert.Equal(t, "model", m.Object)
	}
}

func TestListOpenAIJSON(t *testing.T) {
	b, err := ListOpenAIJSON()
	require.NoError(t, err)
	var list OpenAIModelList
	require.NoError(t, json.Unmarshal(b, &list))
	assert.Equal(t, "list", list.Object)
}
