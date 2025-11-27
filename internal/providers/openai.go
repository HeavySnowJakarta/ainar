// Package providers implements AI provider interfaces
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

// OpenAIProvider implements the Provider interface for OpenAI-compatible APIs
type OpenAIProvider struct {
	name    string
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(baseURL, apiKey string) (Provider, error) {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	// Ensure baseURL doesn't end with /
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &OpenAIProvider{
		name:    "OpenAI",
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}, nil
}

func (p *OpenAIProvider) Name() string {
	return p.name
}

func (p *OpenAIProvider) Type() string {
	return "openai"
}

func (p *OpenAIProvider) SupportedModalities() []Modality {
	return []Modality{ModalityText, ModalityImage, ModalityAudio}
}

func (p *OpenAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Convert internal request to OpenAI format
	openaiReq := p.toOpenAIRequest(req)

	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
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

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}

func (p *OpenAIProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, <-chan error) {
	respChan := make(chan *ChatResponse, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		// Enable streaming
		req.Stream = true
		openaiReq := p.toOpenAIRequest(req)

		body, err := json.Marshal(openaiReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
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

			// Keep the last incomplete line for next iteration
			remainder = lines[len(lines)-1]
			lines = lines[:len(lines)-1]

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				if !strings.HasPrefix(line, "data: ") {
					continue
				}

				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					return
				}

				var chatResp ChatResponse
				if err := json.Unmarshal([]byte(data), &chatResp); err != nil {
					continue
				}

				select {
				case respChan <- &chatResp:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return respChan, errChan
}

func (p *OpenAIProvider) Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/embeddings", bytes.NewReader(body))
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

	var embResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &embResp, nil
}

func (p *OpenAIProvider) ListModels(ctx context.Context) (*ModelsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
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

	var modelsResp ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &modelsResp, nil
}

func (p *OpenAIProvider) Close() error {
	return nil
}

func (p *OpenAIProvider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
}

// toOpenAIRequest converts internal request to OpenAI format
func (p *OpenAIProvider) toOpenAIRequest(req *ChatRequest) map[string]interface{} {
	openaiReq := map[string]interface{}{
		"model":    req.Model,
		"messages": p.convertMessages(req.Messages),
		"stream":   req.Stream,
	}

	if req.Temperature != nil {
		openaiReq["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		openaiReq["top_p"] = *req.TopP
	}
	if req.MaxTokens != nil {
		openaiReq["max_tokens"] = *req.MaxTokens
	}
	if len(req.Stop) > 0 {
		openaiReq["stop"] = req.Stop
	}
	if req.PresencePenalty != nil {
		openaiReq["presence_penalty"] = *req.PresencePenalty
	}
	if req.FrequencyPenalty != nil {
		openaiReq["frequency_penalty"] = *req.FrequencyPenalty
	}
	if req.User != "" {
		openaiReq["user"] = req.User
	}
	if len(req.Tools) > 0 {
		openaiReq["tools"] = req.Tools
	}

	return openaiReq
}

// convertMessages converts internal messages to OpenAI format
func (p *OpenAIProvider) convertMessages(messages []Message) []map[string]interface{} {
	result := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		m := map[string]interface{}{
			"role": msg.Role,
		}

		// Handle content
		if len(msg.Content) == 1 && msg.Content[0].Type == "text" {
			m["content"] = msg.Content[0].Text
		} else if len(msg.Content) > 0 {
			contents := make([]map[string]interface{}, len(msg.Content))
			for j, c := range msg.Content {
				content := map[string]interface{}{
					"type": c.Type,
				}
				switch c.Type {
				case "text":
					content["text"] = c.Text
				case "image_url":
					if c.ImageURL != nil {
						content["image_url"] = map[string]interface{}{
							"url":    c.ImageURL.URL,
							"detail": c.ImageURL.Detail,
						}
					}
				case "audio_url":
					if c.AudioURL != nil {
						content["audio_url"] = map[string]interface{}{
							"url":    c.AudioURL.URL,
							"format": c.AudioURL.Format,
						}
					}
				}
				contents[j] = content
			}
			m["content"] = contents
		}

		result[i] = m
	}
	return result
}

func init() {
	Register("openai", NewOpenAIProvider)
}
