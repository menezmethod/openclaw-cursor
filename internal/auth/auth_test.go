package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripANSI(t *testing.T) {
	s := "\x1b[31mred\x1b[0m normal"
	assert.Equal(t, "red normal", stripANSI(s))
}

func TestAuthPaths(t *testing.T) {
	paths := authPaths()
	assert.NotEmpty(t, paths)
	assert.Contains(t, paths[0], ".cursor")
}

func TestVerifyAuth(t *testing.T) {
	status := VerifyAuth()
	// Status may be true if user has cursor credentials
	_ = status.Authenticated
	_ = status.CredentialPath
}
