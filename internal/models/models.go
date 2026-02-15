package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Model represents a Cursor model.
type Model struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	SupportsThinking bool   `json:"supports_thinking"`
	SupportsTools    bool   `json:"supports_tools"`
}

// Registry contains all supported Cursor models.
var Registry = map[string]Model{
	// Composer
	"auto":         {ID: "auto", Name: "Auto", SupportsThinking: false, SupportsTools: true},
	"composer-1.5": {ID: "composer-1.5", Name: "Composer 1.5", SupportsThinking: false, SupportsTools: true},
	"composer-1":   {ID: "composer-1", Name: "Composer 1", SupportsThinking: false, SupportsTools: true},

	// GPT-5.3 Codex
	"gpt-5.3-codex":      {ID: "gpt-5.3-codex", Name: "GPT-5.3 Codex", SupportsThinking: true, SupportsTools: true},
	"gpt-5.3-codex-low":  {ID: "gpt-5.3-codex-low", Name: "GPT-5.3 Codex Low", SupportsThinking: true, SupportsTools: true},
	"gpt-5.3-codex-high": {ID: "gpt-5.3-codex-high", Name: "GPT-5.3 Codex High", SupportsThinking: true, SupportsTools: true},
	"gpt-5.3-codex-xhigh": {ID: "gpt-5.3-codex-xhigh", Name: "GPT-5.3 Codex Extra High", SupportsThinking: true, SupportsTools: true},
	"gpt-5.3-codex-fast":  {ID: "gpt-5.3-codex-fast", Name: "GPT-5.3 Codex Fast", SupportsThinking: true, SupportsTools: true},
	"gpt-5.3-codex-low-fast":  {ID: "gpt-5.3-codex-low-fast", Name: "GPT-5.3 Codex Low Fast", SupportsThinking: true, SupportsTools: true},
	"gpt-5.3-codex-high-fast": {ID: "gpt-5.3-codex-high-fast", Name: "GPT-5.3 Codex High Fast", SupportsThinking: true, SupportsTools: true},
	"gpt-5.3-codex-xhigh-fast": {ID: "gpt-5.3-codex-xhigh-fast", Name: "GPT-5.3 Codex Extra High Fast", SupportsThinking: true, SupportsTools: true},

	// GPT-5.2
	"gpt-5.2":       {ID: "gpt-5.2", Name: "GPT-5.2", SupportsThinking: true, SupportsTools: true},
	"gpt-5.2-codex": {ID: "gpt-5.2-codex", Name: "GPT-5.2 Codex", SupportsThinking: true, SupportsTools: true},
	"gpt-5.2-codex-high":  {ID: "gpt-5.2-codex-high", Name: "GPT-5.2 Codex High", SupportsThinking: true, SupportsTools: true},
	"gpt-5.2-codex-low":   {ID: "gpt-5.2-codex-low", Name: "GPT-5.2 Codex Low", SupportsThinking: true, SupportsTools: true},
	"gpt-5.2-codex-xhigh": {ID: "gpt-5.2-codex-xhigh", Name: "GPT-5.2 Codex Extra High", SupportsThinking: true, SupportsTools: true},
	"gpt-5.2-codex-fast":       {ID: "gpt-5.2-codex-fast", Name: "GPT-5.2 Codex Fast", SupportsThinking: true, SupportsTools: true},
	"gpt-5.2-codex-high-fast":  {ID: "gpt-5.2-codex-high-fast", Name: "GPT-5.2 Codex High Fast", SupportsThinking: true, SupportsTools: true},
	"gpt-5.2-codex-low-fast":   {ID: "gpt-5.2-codex-low-fast", Name: "GPT-5.2 Codex Low Fast", SupportsThinking: true, SupportsTools: true},
	"gpt-5.2-codex-xhigh-fast": {ID: "gpt-5.2-codex-xhigh-fast", Name: "GPT-5.2 Codex Extra High Fast", SupportsThinking: true, SupportsTools: true},
	"gpt-5.2-high": {ID: "gpt-5.2-high", Name: "GPT-5.2 High", SupportsThinking: true, SupportsTools: true},

	// GPT-5.1
	"gpt-5.1-codex-max":    {ID: "gpt-5.1-codex-max", Name: "GPT-5.1 Codex Max", SupportsThinking: true, SupportsTools: true},
	"gpt-5.1-codex-max-high": {ID: "gpt-5.1-codex-max-high", Name: "GPT-5.1 Codex Max High", SupportsThinking: true, SupportsTools: true},
	"gpt-5.1-high":         {ID: "gpt-5.1-high", Name: "GPT-5.1 High", SupportsThinking: true, SupportsTools: true},

	// Claude
	"opus-4.6":          {ID: "opus-4.6", Name: "Claude 4.6 Opus", SupportsThinking: false, SupportsTools: true},
	"opus-4.6-thinking": {ID: "opus-4.6-thinking", Name: "Claude 4.6 Opus (Thinking)", SupportsThinking: true, SupportsTools: true},
	"opus-4.5":          {ID: "opus-4.5", Name: "Claude 4.5 Opus", SupportsThinking: false, SupportsTools: true},
	"opus-4.5-thinking": {ID: "opus-4.5-thinking", Name: "Claude 4.5 Opus (Thinking)", SupportsThinking: true, SupportsTools: true},
	"sonnet-4.5":          {ID: "sonnet-4.5", Name: "Claude 4.5 Sonnet", SupportsThinking: false, SupportsTools: true},
	"sonnet-4.5-thinking": {ID: "sonnet-4.5-thinking", Name: "Claude 4.5 Sonnet (Thinking)", SupportsThinking: true, SupportsTools: true},

	// Other
	"gemini-3-pro":   {ID: "gemini-3-pro", Name: "Gemini 3 Pro", SupportsThinking: false, SupportsTools: true},
	"gemini-3-flash": {ID: "gemini-3-flash", Name: "Gemini 3 Flash", SupportsThinking: false, SupportsTools: true},
	"grok":           {ID: "grok", Name: "Grok", SupportsThinking: false, SupportsTools: true},
}

// Resolve strips cursor/ or cursor-acp/ prefix and validates model ID.
func Resolve(input string) (string, error) {
	s := strings.TrimSpace(input)
	s = strings.TrimPrefix(s, "cursor/")
	s = strings.TrimPrefix(s, "cursor-acp/")
	if m, ok := Registry[s]; ok {
		return m.ID, nil
	}
	return "", fmt.Errorf("unknown model %q", input)
}

// OpenAIModel represents a model in OpenAI /v1/models response.
type OpenAIModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// OpenAIModelList is the response for GET /v1/models.
type OpenAIModelList struct {
	Object string         `json:"object"`
	Data   []OpenAIModel  `json:"data"`
}

// ListOpenAI returns models in OpenAI API format.
func ListOpenAI() OpenAIModelList {
	data := make([]OpenAIModel, 0, len(Registry))
	for _, m := range Registry {
		data = append(data, OpenAIModel{
			ID:      m.ID,
			Object:  "model",
			Created: 1700000000,
			OwnedBy: "cursor",
		})
	}
	return OpenAIModelList{Object: "list", Data: data}
}

// ListOpenAIJSON returns the JSON bytes for GET /v1/models.
func ListOpenAIJSON() ([]byte, error) {
	return json.Marshal(ListOpenAI())
}
