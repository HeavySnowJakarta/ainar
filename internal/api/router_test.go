package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HeavySnowJakarta/ainar/internal/config"
)

func TestRouterWithoutAuth(t *testing.T) {
	cfg := config.DefaultConfig()
	router := NewRouter(cfg, "")

	// Request without API key
	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["error"] == nil {
		t.Error("Expected error in response")
	}
}

func TestRouterWithValidAuth(t *testing.T) {
	cfg := config.DefaultConfig()
	key, _ := cfg.GenerateDownstreamKey("test-app", nil)
	router := NewRouter(cfg, "")

	// Request with valid API key but invalid model format
	body := map[string]interface{}{
		"model": "gpt-4", // Missing provider prefix
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should fail because model format is invalid (no provider prefix)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid model format, got %d", w.Code)
	}
}

func TestRouterCORS(t *testing.T) {
	cfg := config.DefaultConfig()
	router := NewRouter(cfg, "")

	req := httptest.NewRequest("OPTIONS", "/v1/chat/completions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS header to be set")
	}
}

func TestRouterNotFound(t *testing.T) {
	cfg := config.DefaultConfig()
	key, _ := cfg.GenerateDownstreamKey("test-app", nil)
	router := NewRouter(cfg, "")

	req := httptest.NewRequest("GET", "/v1/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+key)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestParseModelName(t *testing.T) {
	testCases := []struct {
		model        string
		providerCode string
		modelName    string
		expectError  bool
	}{
		{"openai/gpt-4", "openai", "gpt-4", false},
		{"anthropic/claude-3-sonnet", "anthropic", "claude-3-sonnet", false},
		{"openrouter/anthropic/claude-3", "openrouter", "anthropic/claude-3", false},
		{"gpt-4", "", "", true}, // No provider prefix
	}

	for _, tc := range testCases {
		providerCode, modelName, err := parseModelName(tc.model)
		if tc.expectError {
			if err == nil {
				t.Errorf("Expected error for model '%s'", tc.model)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for model '%s': %v", tc.model, err)
			}
			if providerCode != tc.providerCode {
				t.Errorf("Expected provider '%s', got '%s'", tc.providerCode, providerCode)
			}
			if modelName != tc.modelName {
				t.Errorf("Expected model name '%s', got '%s'", tc.modelName, modelName)
			}
		}
	}
}

func TestExtractAPIKey(t *testing.T) {
	testCases := []struct {
		header   string
		expected string
	}{
		{"Bearer sk-1234567890", "sk-1234567890"},
		{"Bearer ainar-abc123def456", "ainar-abc123def456"},
		{"", ""},
		{"Basic dXNlcjpwYXNz", ""}, // Basic auth should return empty
	}

	for _, tc := range testCases {
		req := httptest.NewRequest("GET", "/", nil)
		if tc.header != "" {
			req.Header.Set("Authorization", tc.header)
		}

		key := extractAPIKey(req)
		if key != tc.expected {
			t.Errorf("For header '%s', expected key '%s', got '%s'", tc.header, tc.expected, key)
		}
	}
}

func TestRouterListModels(t *testing.T) {
	cfg := config.DefaultConfig()
	key, _ := cfg.GenerateDownstreamKey("test-app", nil)
	router := NewRouter(cfg, "")

	req := httptest.NewRequest("GET", "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+key)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["object"] != "list" {
		t.Error("Expected object to be 'list'")
	}
}
