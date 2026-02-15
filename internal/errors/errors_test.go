package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_Quota(t *testing.T) {
	pe := Parse("Error: You have exceeded your usage limit")
	assert.Equal(t, "quota_exceeded", pe.Type)
	assert.False(t, pe.Recoverable)
}

func TestParse_Auth(t *testing.T) {
	pe := Parse("Authentication failed: not logged in")
	assert.Equal(t, "auth_failed", pe.Type)
	assert.Contains(t, pe.Suggestion, "openclaw-cursor login")
}

func TestParse_Model(t *testing.T) {
	pe := Parse("model not found: xyz")
	assert.Equal(t, "model_unavailable", pe.Type)
}

func TestParse_Network(t *testing.T) {
	pe := Parse("fetch failed: ECONNREFUSED")
	assert.Equal(t, "network_error", pe.Type)
	assert.True(t, pe.Recoverable)
}

func TestToOpenAIErrorJSON(t *testing.T) {
	pe := Parse("quota exceeded")
	b, err := ToOpenAIErrorJSON(pe)
	require.NoError(t, err)
	assert.Contains(t, string(b), "quota_exceeded")
	assert.Contains(t, string(b), "error")
}
