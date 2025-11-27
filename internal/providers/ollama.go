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

// OllamaProvider implements the Provider interface for Ollama/LM Studio
type OllamaProvider struct {
	name    string
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL, apiKey string) (Provider, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &OllamaProvider{
		name:    "Ollama",
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}, nil
}

func (p *OllamaProvider) Name() string {
	return p.name
}

func (p *OllamaProvider) Type() string {
	return "ollama"
}

func (p *OllamaProvider) SupportedModalities() []Modality {
	return []Modality{ModalityText, ModalityImage}
}

func (p *OllamaProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	ollamaReq := p.toOllamaRequest(req)

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.toOpenAIResponse(&ollamaResp, req.Model), nil
}

func (p *OllamaProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, <-chan error) {
	respChan := make(chan *ChatResponse, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		ollamaReq := p.toOllamaRequest(req)
		ollamaReq["stream"] = true

		body, err := json.Marshal(ollamaReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(body))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		if p.apiKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
		}

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

		decoder := json.NewDecoder(resp.Body)
		for {
			var ollamaResp ollamaResponse
			if err := decoder.Decode(&ollamaResp); err != nil {
				if err == io.EOF {
					return
				}
				errChan <- fmt.Errorf("failed to decode response: %w", err)
				return
			}

			chatResp := &ChatResponse{
				ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   req.Model,
				Choices: []Choice{
					{
						Index: 0,
						Delta: &ChatMessage{
							Content: ollamaResp.Message.Content,
						},
					},
				},
			}

			if ollamaResp.Done {
				chatResp.Choices[0].FinishReason = "stop"
			}

			select {
			case respChan <- chatResp:
			case <-ctx.Done():
				return
			}

			if ollamaResp.Done {
				return
			}
		}
	}()

	return respChan, errChan
}

func (p *OllamaProvider) Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	ollamaReq := map[string]interface{}{
		"model":  req.Model,
		"prompt": strings.Join(req.Input, " "),
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &EmbeddingResponse{
		Object: "list",
		Data: []EmbeddingData{
			{
				Object:    "embedding",
				Embedding: ollamaResp.Embedding,
				Index:     0,
			},
		},
		Model: req.Model,
	}, nil
}

func (p *OllamaProvider) ListModels(ctx context.Context) (*ModelsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp struct {
		Models []struct {
			Name       string    `json:"name"`
			ModifiedAt time.Time `json:"modified_at"`
			Size       int64     `json:"size"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]ModelInfo, len(ollamaResp.Models))
	for i, m := range ollamaResp.Models {
		models[i] = ModelInfo{
			ID:      m.Name,
			Object:  "model",
			Created: m.ModifiedAt.Unix(),
			OwnedBy: "ollama",
		}
	}

	return &ModelsResponse{
		Object: "list",
		Data:   models,
	}, nil
}

func (p *OllamaProvider) Close() error {
	return nil
}

// toOllamaRequest converts internal request to Ollama format
func (p *OllamaProvider) toOllamaRequest(req *ChatRequest) map[string]interface{} {
	ollamaReq := map[string]interface{}{
		"model":  req.Model,
		"stream": false,
	}

	// Convert messages
	messages := make([]map[string]interface{}, len(req.Messages))
	for i, msg := range req.Messages {
		m := map[string]interface{}{
			"role": msg.Role,
		}

		// Handle content
		if len(msg.Content) == 1 && msg.Content[0].Type == "text" {
			m["content"] = msg.Content[0].Text
		} else if len(msg.Content) > 0 {
			var text string
			var images []string
			for _, c := range msg.Content {
				switch c.Type {
				case "text":
					text += c.Text
				case "image_url":
					if c.ImageURL != nil {
						images = append(images, c.ImageURL.URL)
					}
				}
			}
			m["content"] = text
			if len(images) > 0 {
				m["images"] = images
			}
		}

		messages[i] = m
	}
	ollamaReq["messages"] = messages

	// Options
	options := map[string]interface{}{}
	if req.Temperature != nil {
		options["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		options["top_p"] = *req.TopP
	}
	if req.MaxTokens != nil {
		options["num_predict"] = *req.MaxTokens
	}
	if len(req.Stop) > 0 {
		options["stop"] = req.Stop
	}

	if len(options) > 0 {
		ollamaReq["options"] = options
	}

	return ollamaReq
}

// ollamaResponse represents Ollama's response format
type ollamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Message   struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done               bool  `json:"done"`
	TotalDuration      int64 `json:"total_duration"`
	LoadDuration       int64 `json:"load_duration"`
	PromptEvalCount    int   `json:"prompt_eval_count"`
	PromptEvalDuration int64 `json:"prompt_eval_duration"`
	EvalCount          int   `json:"eval_count"`
	EvalDuration       int64 `json:"eval_duration"`
}

// toOpenAIResponse converts Ollama response to OpenAI format
func (p *OllamaProvider) toOpenAIResponse(resp *ollamaResponse, model string) *ChatResponse {
	return &ChatResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []Choice{
			{
				Index: 0,
				Message: &ChatMessage{
					Role:    resp.Message.Role,
					Content: resp.Message.Content,
				},
				FinishReason: "stop",
			},
		},
		Usage: &Usage{
			PromptTokens:     resp.PromptEvalCount,
			CompletionTokens: resp.EvalCount,
			TotalTokens:      resp.PromptEvalCount + resp.EvalCount,
		},
	}
}

func init() {
	Register("ollama", NewOllamaProvider)
}
