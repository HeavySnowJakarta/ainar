package providers

import (
	"testing"
)

func TestKnownProviders(t *testing.T) {
	providers := KnownProviders()

	if len(providers) == 0 {
		t.Error("Expected at least one known provider")
	}

	// Check that OpenAI is in the list
	found := false
	for _, p := range providers {
		if p.Code == "openai" {
			found = true
			if p.APIType != "openai" {
				t.Errorf("OpenAI should have api_type 'openai', got '%s'", p.APIType)
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find OpenAI in known providers")
	}
}

func TestFindKnownProvider(t *testing.T) {
	// Search by name
	results := FindKnownProvider("openai")
	if len(results) == 0 {
		t.Error("Expected to find providers matching 'openai'")
	}

	// Search case-insensitive
	results = FindKnownProvider("OPENAI")
	if len(results) == 0 {
		t.Error("Expected case-insensitive search to work")
	}

	// Search partial
	results = FindKnownProvider("open")
	if len(results) == 0 {
		t.Error("Expected partial search to work")
	}

	// Non-existent
	results = FindKnownProvider("nonexistent-provider-xyz")
	if len(results) != 0 {
		t.Error("Expected no results for non-existent provider")
	}
}

func TestFindKnownProviderByAPIType(t *testing.T) {
	// OpenAI compatible
	results := FindKnownProviderByAPIType("openai")
	if len(results) == 0 {
		t.Error("Expected to find OpenAI-compatible providers")
	}

	// All should have the right API type
	for _, p := range results {
		if p.APIType != "openai" {
			t.Errorf("Expected api_type 'openai', got '%s'", p.APIType)
		}
	}

	// Anthropic
	results = FindKnownProviderByAPIType("anthropic")
	if len(results) == 0 {
		t.Error("Expected to find Anthropic providers")
	}

	// Non-existent
	results = FindKnownProviderByAPIType("nonexistent")
	if len(results) != 0 {
		t.Error("Expected no results for non-existent API type")
	}
}

func TestProviderRegistry(t *testing.T) {
	// Check that providers are registered
	registered := ListRegistered()
	if len(registered) == 0 {
		t.Error("Expected at least one registered provider type")
	}

	// Check for expected types
	expectedTypes := []string{"openai", "anthropic", "google", "ollama"}
	for _, expected := range expectedTypes {
		found := false
		for _, r := range registered {
			if r == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected provider type '%s' to be registered", expected)
		}
	}
}

func TestCreateProvider(t *testing.T) {
	// Create OpenAI provider
	p, err := Create("openai", "https://api.openai.com/v1", "test-key")
	if err != nil {
		t.Fatalf("Failed to create OpenAI provider: %v", err)
	}

	if p.Type() != "openai" {
		t.Errorf("Expected type 'openai', got '%s'", p.Type())
	}

	// Create Anthropic provider
	p, err = Create("anthropic", "https://api.anthropic.com", "test-key")
	if err != nil {
		t.Fatalf("Failed to create Anthropic provider: %v", err)
	}

	if p.Type() != "anthropic" {
		t.Errorf("Expected type 'anthropic', got '%s'", p.Type())
	}

	// Create with unsupported type
	_, err = Create("unsupported", "http://localhost", "key")
	if err == nil {
		t.Error("Expected error for unsupported provider type")
	}
}

func TestSupportedModalities(t *testing.T) {
	testCases := []struct {
		providerType string
		expected     []Modality
	}{
		{"openai", []Modality{ModalityText, ModalityImage, ModalityAudio}},
		{"anthropic", []Modality{ModalityText, ModalityImage}},
		{"google", []Modality{ModalityText, ModalityImage, ModalityAudio, ModalityVideo}},
		{"ollama", []Modality{ModalityText, ModalityImage}},
	}

	for _, tc := range testCases {
		p, err := Create(tc.providerType, "http://localhost", "test-key")
		if err != nil {
			t.Fatalf("Failed to create %s provider: %v", tc.providerType, err)
		}

		modalities := p.SupportedModalities()
		if len(modalities) != len(tc.expected) {
			t.Errorf("%s: Expected %d modalities, got %d", tc.providerType, len(tc.expected), len(modalities))
		}
	}
}
