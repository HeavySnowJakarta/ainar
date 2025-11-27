// Package providers defines interfaces and implementations for AI providers
package providers

import (
	"context"
	"io"
)

// Modality represents the type of content that can be processed
type Modality string

const (
	ModalityText  Modality = "text"
	ModalityImage Modality = "image"
	ModalityAudio Modality = "audio"
	ModalityVideo Modality = "video"
)

// Message represents a chat message
type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

// Content represents content within a message
type Content struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
	AudioURL *AudioURL `json:"audio_url,omitempty"`
	VideoURL *VideoURL `json:"video_url,omitempty"`
}

// ImageURL represents an image in a message
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// AudioURL represents audio in a message
type AudioURL struct {
	URL    string `json:"url"`
	Format string `json:"format,omitempty"`
}

// VideoURL represents video in a message
type VideoURL struct {
	URL    string `json:"url"`
	Format string `json:"format,omitempty"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Temperature      *float64  `json:"temperature,omitempty"`
	TopP             *float64  `json:"top_p,omitempty"`
	MaxTokens        *int      `json:"max_tokens,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
	Stop             []string  `json:"stop,omitempty"`
	PresencePenalty  *float64  `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64  `json:"frequency_penalty,omitempty"`
	User             string    `json:"user,omitempty"`
	// Tools for function calling
	Tools []Tool `json:"tools,omitempty"`
	// Additional provider-specific options
	Extra map[string]interface{} `json:"-"`
}

// Tool represents a tool/function that can be called
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents a function definition
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
}

// Choice represents a response choice
type Choice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message,omitempty"`
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason string       `json:"finish_reason,omitempty"`
}

// ChatMessage represents a message in a chat response
type ChatMessage struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call in a response
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  *Usage          `json:"usage,omitempty"`
}

// EmbeddingData represents a single embedding
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// ModelInfo represents information about a model
type ModelInfo struct {
	ID         string     `json:"id"`
	Object     string     `json:"object"`
	Created    int64      `json:"created"`
	OwnedBy    string     `json:"owned_by"`
	Modalities []Modality `json:"modalities,omitempty"`
}

// ModelsResponse represents a list models response
type ModelsResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}

// Provider defines the interface for AI providers
type Provider interface {
	// Name returns the provider's name
	Name() string

	// Type returns the API type (e.g., "openai", "anthropic")
	Type() string

	// SupportedModalities returns the modalities this provider supports
	SupportedModalities() []Modality

	// Chat performs a chat completion request
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream performs a streaming chat completion request
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, <-chan error)

	// Embedding generates embeddings for the given input
	Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// ListModels lists available models
	ListModels(ctx context.Context) (*ModelsResponse, error)

	// Close closes any resources held by the provider
	Close() error
}

// StreamWriter is an interface for writing streaming responses
type StreamWriter interface {
	io.Writer
	Flush() error
}

// ProviderFactory is a function that creates a provider
type ProviderFactory func(baseURL, apiKey string) (Provider, error)

// registry holds registered provider factories
var registry = make(map[string]ProviderFactory)

// Register registers a provider factory
func Register(apiType string, factory ProviderFactory) {
	registry[apiType] = factory
}

// Create creates a provider using a registered factory
func Create(apiType, baseURL, apiKey string) (Provider, error) {
	factory, ok := registry[apiType]
	if !ok {
		return nil, &ErrUnsupportedProvider{Type: apiType}
	}
	return factory(baseURL, apiKey)
}

// ListRegistered returns a list of registered provider types
func ListRegistered() []string {
	types := make([]string, 0, len(registry))
	for t := range registry {
		types = append(types, t)
	}
	return types
}

// ErrUnsupportedProvider indicates an unsupported provider type
type ErrUnsupportedProvider struct {
	Type string
}

func (e *ErrUnsupportedProvider) Error() string {
	return "unsupported provider type: " + e.Type
}
