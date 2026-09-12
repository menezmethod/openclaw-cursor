package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GabrieleRisso/openclaw-cursor/internal/config"
	"github.com/GabrieleRisso/openclaw-cursor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_Health(t *testing.T) {
	cfg := config.Default()
	log := logger.New("info")
	srv := New(cfg, log, "test")
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	srv.handleHealth(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var m map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&m))
	assert.Equal(t, "healthy", m["status"])
}

func TestServer_ListModels(t *testing.T) {
	cfg := config.Default()
	log := logger.New("info")
	srv := New(cfg, log, "test")
	req := httptest.NewRequest("GET", "/v1/models", nil)
	w := httptest.NewRecorder()
	srv.handleListModels(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var m map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&m))
	assert.Equal(t, "list", m["object"])
	data, ok := m["data"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(data), 30)
}

func TestServer_ChatCompletions_InvalidModel(t *testing.T) {
	cfg := config.Default()
	log := logger.New("info")
	srv := New(cfg, log, "test")
	body := []byte(`{"model":"cursor/invalid-model","messages":[{"role":"user","content":"hi"}]}`)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.handleChatCompletions(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	var m map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&m))
	assert.Contains(t, m, "error")
}
