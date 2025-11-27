// Package config handles configuration management for AINAR
package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	mu sync.RWMutex `yaml:"-"`

	// Server settings
	Server ServerConfig `yaml:"server"`

	// WebUI settings
	WebUI WebUIConfig `yaml:"webui"`

	// Upstream providers
	Providers []Provider `yaml:"providers"`

	// Downstream API keys
	DownstreamKeys []DownstreamKey `yaml:"downstream_keys"`

	// Model aliases
	Aliases []Alias `yaml:"aliases"`
}

// ServerConfig contains server-related settings
type ServerConfig struct {
	// Port for the OpenAI-compatible API
	Port int `yaml:"port"`
	// Host to bind to
	Host string `yaml:"host"`
}

// WebUIConfig contains WebUI-related settings
type WebUIConfig struct {
	// Status: "disabled", "local", "external"
	Status string `yaml:"status"`
	// Port for the WebUI
	Port int `yaml:"port"`
}

// Provider represents an upstream AI provider
type Provider struct {
	// Formal name (e.g., "OpenRouter")
	Name string `yaml:"name"`
	// Code used in model names (e.g., "openrouter")
	Code string `yaml:"code"`
	// API type (e.g., "openai", "anthropic", "google", "ollama")
	APIType string `yaml:"api_type"`
	// Base URL for the API
	BaseURL string `yaml:"base_url"`
	// API Key or authentication token
	APIKey string `yaml:"api_key"`
	// OAuth configuration (for providers like Copilot)
	OAuth *OAuthConfig `yaml:"oauth,omitempty"`
	// Whether the provider is enabled
	Enabled bool `yaml:"enabled"`
}

// OAuthConfig contains OAuth-related settings for providers like Copilot
type OAuthConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	AccessToken  string `yaml:"access_token"`
	RefreshToken string `yaml:"refresh_token"`
	ExpiresAt    int64  `yaml:"expires_at"`
}

// DownstreamKey represents an API key for downstream applications
type DownstreamKey struct {
	// Name of the key
	Name string `yaml:"name"`
	// The actual API key
	Key string `yaml:"key"`
	// Token limit configuration
	Limit *TokenLimit `yaml:"limit,omitempty"`
	// Usage tracking
	Usage *TokenUsage `yaml:"usage,omitempty"`
	// Created timestamp
	CreatedAt time.Time `yaml:"created_at"`
	// Whether the key is enabled
	Enabled bool `yaml:"enabled"`
}

// TokenLimit defines token usage limits
type TokenLimit struct {
	// Type: "daily", "weekly", "monthly", "total"
	Type string `yaml:"type"`
	// Maximum tokens allowed
	MaxTokens int64 `yaml:"max_tokens"`
}

// TokenUsage tracks token usage
type TokenUsage struct {
	// Tokens used
	TokensUsed int64 `yaml:"tokens_used"`
	// Last reset time
	LastReset time.Time `yaml:"last_reset"`
}

// Alias represents a model name alias
type Alias struct {
	// The alias (e.g., "ora")
	From string `yaml:"from"`
	// The actual prefix (e.g., "openrouter/anthropic")
	To string `yaml:"to"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "127.0.0.1",
		},
		WebUI: WebUIConfig{
			Status: "local",
			Port:   8081,
		},
		Providers:      []Provider{},
		DownstreamKeys: []DownstreamKey{},
		Aliases:        []Alias{},
	}
}

// Load loads configuration from a file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to a file
func (c *Config) Save(path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddProvider adds a new provider
func (c *Config) AddProvider(p Provider) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate provider code
	if !isValidCode(p.Code) {
		return fmt.Errorf("invalid provider code: %s (only lowercase letters, numbers, '.', '-' and '_' allowed)", p.Code)
	}

	// Check for duplicate code
	for _, existing := range c.Providers {
		if existing.Code == p.Code {
			return fmt.Errorf("provider with code %s already exists", p.Code)
		}
	}

	c.Providers = append(c.Providers, p)
	return nil
}

// RemoveProvider removes a provider by code
func (c *Config) RemoveProvider(code string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, p := range c.Providers {
		if p.Code == code {
			c.Providers = append(c.Providers[:i], c.Providers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("provider with code %s not found", code)
}

// GetProvider returns a provider by code
func (c *Config) GetProvider(code string) (*Provider, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for i := range c.Providers {
		if c.Providers[i].Code == code {
			return &c.Providers[i], nil
		}
	}
	return nil, fmt.Errorf("provider with code %s not found", code)
}

// GenerateDownstreamKey generates a new API key for downstream applications
func (c *Config) GenerateDownstreamKey(name string, limit *TokenLimit) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Generate a secure random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	key := "ainar-" + hex.EncodeToString(keyBytes)

	dk := DownstreamKey{
		Name:      name,
		Key:       key,
		Limit:     limit,
		Usage:     &TokenUsage{LastReset: time.Now()},
		CreatedAt: time.Now(),
		Enabled:   true,
	}

	c.DownstreamKeys = append(c.DownstreamKeys, dk)
	return key, nil
}

// ValidateDownstreamKey validates a downstream API key
func (c *Config) ValidateDownstreamKey(key string) (*DownstreamKey, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for i := range c.DownstreamKeys {
		if c.DownstreamKeys[i].Key == key && c.DownstreamKeys[i].Enabled {
			return &c.DownstreamKeys[i], nil
		}
	}
	return nil, fmt.Errorf("invalid API key")
}

// AddAlias adds a new model alias
func (c *Config) AddAlias(from, to string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check for duplicate
	for _, a := range c.Aliases {
		if a.From == from {
			return fmt.Errorf("alias %s already exists", from)
		}
	}

	c.Aliases = append(c.Aliases, Alias{From: from, To: to})
	return nil
}

// ResolveAlias resolves a model name using aliases
func (c *Config) ResolveAlias(model string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, a := range c.Aliases {
		if model == a.From || len(model) > len(a.From) && model[:len(a.From)+1] == a.From+"/" {
			return a.To + model[len(a.From):]
		}
	}
	return model
}

// isValidCode checks if a provider code is valid
func isValidCode(code string) bool {
	if code == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-z0-9._-]+$`, code)
	return matched
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	// Check for config in current directory first
	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml"
	}

	// Then check XDG config directory
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "config.yaml"
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configDir, "ainar", "config.yaml")
}
