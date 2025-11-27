package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default server port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Expected default host 127.0.0.1, got %s", cfg.Server.Host)
	}

	if cfg.WebUI.Status != "local" {
		t.Errorf("Expected default WebUI status 'local', got %s", cfg.WebUI.Status)
	}

	if cfg.WebUI.Port != 8081 {
		t.Errorf("Expected default WebUI port 8081, got %d", cfg.WebUI.Port)
	}
}

func TestProviderValidation(t *testing.T) {
	cfg := DefaultConfig()

	// Valid codes
	validCodes := []string{"openai", "my-provider", "provider_1", "test.provider"}
	for _, code := range validCodes {
		p := Provider{Name: "Test", Code: code, APIType: "openai", BaseURL: "http://localhost", Enabled: true}
		if err := cfg.AddProvider(p); err != nil {
			t.Errorf("Expected code '%s' to be valid, got error: %v", code, err)
		}
		cfg.RemoveProvider(code)
	}

	// Invalid codes
	invalidCodes := []string{"OpenAI", "my provider", "test@provider", ""}
	for _, code := range invalidCodes {
		p := Provider{Name: "Test", Code: code, APIType: "openai", BaseURL: "http://localhost", Enabled: true}
		if err := cfg.AddProvider(p); err == nil {
			t.Errorf("Expected code '%s' to be invalid", code)
		}
	}
}

func TestProviderDuplicate(t *testing.T) {
	cfg := DefaultConfig()

	p1 := Provider{Name: "Test 1", Code: "test", APIType: "openai", BaseURL: "http://localhost", Enabled: true}
	if err := cfg.AddProvider(p1); err != nil {
		t.Fatalf("Failed to add first provider: %v", err)
	}

	p2 := Provider{Name: "Test 2", Code: "test", APIType: "openai", BaseURL: "http://localhost2", Enabled: true}
	if err := cfg.AddProvider(p2); err == nil {
		t.Error("Expected error for duplicate provider code")
	}
}

func TestDownstreamKey(t *testing.T) {
	cfg := DefaultConfig()

	key, err := cfg.GenerateDownstreamKey("test-app", nil)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	if key == "" {
		t.Error("Generated key should not be empty")
	}

	if len(key) < 32 {
		t.Error("Generated key should be at least 32 characters")
	}

	// Validate the key
	dk, err := cfg.ValidateDownstreamKey(key)
	if err != nil {
		t.Errorf("Failed to validate key: %v", err)
	}

	if dk.Name != "test-app" {
		t.Errorf("Expected key name 'test-app', got %s", dk.Name)
	}

	// Invalid key
	_, err = cfg.ValidateDownstreamKey("invalid-key")
	if err == nil {
		t.Error("Expected error for invalid key")
	}
}

func TestAlias(t *testing.T) {
	cfg := DefaultConfig()

	if err := cfg.AddAlias("ora", "openrouter/anthropic"); err != nil {
		t.Fatalf("Failed to add alias: %v", err)
	}

	// Test resolution
	resolved := cfg.ResolveAlias("ora/claude-3")
	if resolved != "openrouter/anthropic/claude-3" {
		t.Errorf("Expected 'openrouter/anthropic/claude-3', got '%s'", resolved)
	}

	// Test exact match
	resolved = cfg.ResolveAlias("ora")
	if resolved != "openrouter/anthropic" {
		t.Errorf("Expected 'openrouter/anthropic', got '%s'", resolved)
	}

	// Test no match
	resolved = cfg.ResolveAlias("other/model")
	if resolved != "other/model" {
		t.Errorf("Expected 'other/model', got '%s'", resolved)
	}

	// Duplicate alias
	if err := cfg.AddAlias("ora", "something-else"); err == nil {
		t.Error("Expected error for duplicate alias")
	}
}

func TestConfigSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create and save config
	cfg := DefaultConfig()
	cfg.AddProvider(Provider{
		Name:    "Test Provider",
		Code:    "test",
		APIType: "openai",
		BaseURL: "http://localhost:8000",
		APIKey:  "test-key",
		Enabled: true,
	})
	cfg.AddAlias("t", "test")

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(loaded.Providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(loaded.Providers))
	}

	if loaded.Providers[0].Code != "test" {
		t.Errorf("Expected provider code 'test', got '%s'", loaded.Providers[0].Code)
	}

	if len(loaded.Aliases) != 1 {
		t.Errorf("Expected 1 alias, got %d", len(loaded.Aliases))
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("Expected no error for non-existent config, got: %v", err)
	}

	// Should return default config
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default server port, got %d", cfg.Server.Port)
	}
}

func TestTokenLimitWithKey(t *testing.T) {
	cfg := DefaultConfig()

	limit := &TokenLimit{
		Type:      "daily",
		MaxTokens: 10000,
	}

	key, err := cfg.GenerateDownstreamKey("limited-app", limit)
	if err != nil {
		t.Fatalf("Failed to generate key with limit: %v", err)
	}

	dk, err := cfg.ValidateDownstreamKey(key)
	if err != nil {
		t.Fatalf("Failed to validate key: %v", err)
	}

	if dk.Limit == nil {
		t.Fatal("Expected limit to be set")
	}

	if dk.Limit.Type != "daily" {
		t.Errorf("Expected limit type 'daily', got '%s'", dk.Limit.Type)
	}

	if dk.Limit.MaxTokens != 10000 {
		t.Errorf("Expected max tokens 10000, got %d", dk.Limit.MaxTokens)
	}
}

func TestGetConfigPath(t *testing.T) {
	// Save original directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Test with config.yaml in current directory
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	os.WriteFile("config.yaml", []byte("test"), 0644)

	path := GetConfigPath()
	if path != "config.yaml" {
		t.Errorf("Expected 'config.yaml' when file exists, got '%s'", path)
	}
}
