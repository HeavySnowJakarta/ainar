package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AnthropicProvider implements the Provider interface for Anthropic's Claude API
type AnthropicProvider struct {
	name    string
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(baseURL, apiKey string) (Provider, error) {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &AnthropicProvider{
		name:    "Anthropic",
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}, nil
}

func (p *AnthropicProvider) Name() string {
	return p.name
}

func (p *AnthropicProvider) Type() string {
	return "anthropic"
}

func (p *AnthropicProvider) SupportedModalities() []Modality {
	return []Modality{ModalityText, ModalityImage}
}

func (p *AnthropicProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	anthropicReq := p.toAnthropicRequest(req)

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	p.setHeaders(httpReq)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.toOpenAIResponse(&anthropicResp), nil
}

func (p *AnthropicProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, <-chan error) {
	respChan := make(chan *ChatResponse, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		anthropicReq := p.toAnthropicRequest(req)
		anthropicReq["stream"] = true

		body, err := json.Marshal(anthropicReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(body))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		p.setHeaders(httpReq)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			errChan <- fmt.Errorf("request failed: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
			return
		}

		// Read SSE stream
		reader := resp.Body
		buf := make([]byte, 4096)
		var remainder string
		var currentID string

		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					errChan <- fmt.Errorf("stream read error: %w", err)
				}
				return
			}

			data := remainder + string(buf[:n])
			lines := strings.Split(data, "\n")
			remainder = lines[len(lines)-1]
			lines = lines[:len(lines)-1]

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "event:") {
					continue
				}

				if !strings.HasPrefix(line, "data: ") {
					continue
				}

				data := strings.TrimPrefix(line, "data: ")

				var event anthropicStreamEvent
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					continue
				}

				switch event.Type {
				case "message_start":
					if event.Message != nil {
						currentID = event.Message.ID
					}
				case "content_block_delta":
					if event.Delta != nil && event.Delta.Text != "" {
						chatResp := &ChatResponse{
							ID:      currentID,
							Object:  "chat.completion.chunk",
							Created: time.Now().Unix(),
							Model:   req.Model,
							Choices: []Choice{
								{
									Index: 0,
									Delta: &ChatMessage{
										Content: event.Delta.Text,
									},
								},
							},
						}
						select {
						case respChan <- chatResp:
						case <-ctx.Done():
							return
						}
					}
				case "message_stop":
					chatResp := &ChatResponse{
						ID:      currentID,
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   req.Model,
						Choices: []Choice{
							{
								Index:        0,
								FinishReason: "stop",
							},
						},
					}
					select {
					case respChan <- chatResp:
					case <-ctx.Done():
						return
					}
					return
				}
			}
		}
	}()

	return respChan, errChan
}

func (p *AnthropicProvider) Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	return nil, fmt.Errorf("anthropic does not support embeddings")
}

func (p *AnthropicProvider) ListModels(ctx context.Context) (*ModelsResponse, error) {
	// Anthropic doesn't have a list models endpoint, return known models
	return &ModelsResponse{
		Object: "list",
		Data: []ModelInfo{
			{ID: "claude-3-5-sonnet-20241022", Object: "model", OwnedBy: "anthropic"},
			{ID: "claude-3-5-haiku-20241022", Object: "model", OwnedBy: "anthropic"},
			{ID: "claude-3-opus-20240229", Object: "model", OwnedBy: "anthropic"},
			{ID: "claude-3-sonnet-20240229", Object: "model", OwnedBy: "anthropic"},
			{ID: "claude-3-haiku-20240307", Object: "model", OwnedBy: "anthropic"},
		},
	}, nil
}

func (p *AnthropicProvider) Close() error {
	return nil
}

func (p *AnthropicProvider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
}

// toAnthropicRequest converts internal request to Anthropic format
func (p *AnthropicProvider) toAnthropicRequest(req *ChatRequest) map[string]interface{} {
	anthropicReq := map[string]interface{}{
		"model": req.Model,
	}

	// Extract system message and convert messages
	var systemMsg string
	messages := make([]map[string]interface{}, 0)

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			if len(msg.Content) > 0 && msg.Content[0].Type == "text" {
				systemMsg = msg.Content[0].Text
			}
			continue
		}

		m := map[string]interface{}{
			"role": msg.Role,
		}

		// Convert content
		if len(msg.Content) == 1 && msg.Content[0].Type == "text" {
			m["content"] = msg.Content[0].Text
		} else if len(msg.Content) > 0 {
			contents := make([]map[string]interface{}, 0)
			for _, c := range msg.Content {
				content := map[string]interface{}{
					"type": c.Type,
				}
				switch c.Type {
				case "text":
					content["text"] = c.Text
				case "image_url":
					if c.ImageURL != nil {
						// Anthropic expects base64 images differently
						content["type"] = "image"
						content["source"] = map[string]interface{}{
							"type": "url",
							"url":  c.ImageURL.URL,
						}
					}
				}
				contents = append(contents, content)
			}
			m["content"] = contents
		}

		messages = append(messages, m)
	}

	anthropicReq["messages"] = messages

	if systemMsg != "" {
		anthropicReq["system"] = systemMsg
	}

	if req.MaxTokens != nil {
		anthropicReq["max_tokens"] = *req.MaxTokens
	} else {
		anthropicReq["max_tokens"] = 4096 // Anthropic requires max_tokens
	}

	if req.Temperature != nil {
		anthropicReq["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		anthropicReq["top_p"] = *req.TopP
	}
	if len(req.Stop) > 0 {
		anthropicReq["stop_sequences"] = req.Stop
	}

	return anthropicReq
}

// anthropicResponse represents Anthropic's response format
type anthropicResponse struct {
	ID           string             `json:"id"`
	Type         string             `json:"type"`
	Role         string             `json:"role"`
	Content      []anthropicContent `json:"content"`
	Model        string             `json:"model"`
	StopReason   string             `json:"stop_reason"`
	StopSequence string             `json:"stop_sequence"`
	Usage        *anthropicUsage    `json:"usage"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicStreamEvent struct {
	Type    string                   `json:"type"`
	Message *anthropicMessageStart   `json:"message,omitempty"`
	Delta   *anthropicContentDelta   `json:"delta,omitempty"`
	Usage   *anthropicUsage          `json:"usage,omitempty"`
}

type anthropicMessageStart struct {
	ID    string `json:"id"`
	Model string `json:"model"`
}

type anthropicContentDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// toOpenAIResponse converts Anthropic response to OpenAI format
func (p *AnthropicProvider) toOpenAIResponse(resp *anthropicResponse) *ChatResponse {
	content := ""
	for _, c := range resp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	chatResp := &ChatResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   resp.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: &ChatMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: p.mapStopReason(resp.StopReason),
			},
		},
	}

	if resp.Usage != nil {
		chatResp.Usage = &Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		}
	}

	return chatResp
}

func (p *AnthropicProvider) mapStopReason(reason string) string {
	switch reason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "stop_sequence":
		return "stop"
	default:
		return reason
	}
}

func init() {
	Register("anthropic", NewAnthropicProvider)
}
