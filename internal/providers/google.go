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

// GoogleProvider implements the Provider interface for Google's Gemini API
type GoogleProvider struct {
	name    string
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewGoogleProvider creates a new Google provider
func NewGoogleProvider(baseURL, apiKey string) (Provider, error) {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &GoogleProvider{
		name:    "Google",
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}, nil
}

func (p *GoogleProvider) Name() string {
	return p.name
}

func (p *GoogleProvider) Type() string {
	return "google"
}

func (p *GoogleProvider) SupportedModalities() []Modality {
	return []Modality{ModalityText, ModalityImage, ModalityAudio, ModalityVideo}
}

func (p *GoogleProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	geminiReq := p.toGeminiRequest(req)

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", p.baseURL, req.Model, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.toOpenAIResponse(&geminiResp, req.Model), nil
}

func (p *GoogleProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, <-chan error) {
	respChan := make(chan *ChatResponse, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		geminiReq := p.toGeminiRequest(req)

		body, err := json.Marshal(geminiReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		url := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", p.baseURL, req.Model, p.apiKey)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")

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

				var geminiResp geminiResponse
				if err := json.Unmarshal([]byte(data), &geminiResp); err != nil {
					continue
				}

				chatResp := p.toOpenAIStreamResponse(&geminiResp, req.Model)
				if chatResp != nil {
					select {
					case respChan <- chatResp:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return respChan, errChan
}

func (p *GoogleProvider) Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	geminiReq := map[string]interface{}{
		"content": map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": strings.Join(req.Input, " ")},
			},
		},
	}

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	model := req.Model
	if model == "" {
		model = "text-embedding-004"
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:embedContent?key=%s", p.baseURL, model, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var geminiResp struct {
		Embedding struct {
			Values []float64 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &EmbeddingResponse{
		Object: "list",
		Data: []EmbeddingData{
			{
				Object:    "embedding",
				Embedding: geminiResp.Embedding.Values,
				Index:     0,
			},
		},
		Model: model,
	}, nil
}

func (p *GoogleProvider) ListModels(ctx context.Context) (*ModelsResponse, error) {
	url := fmt.Sprintf("%s/v1beta/models?key=%s", p.baseURL, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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

	var geminiResp struct {
		Models []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]ModelInfo, len(geminiResp.Models))
	for i, m := range geminiResp.Models {
		// Extract model ID from name (format: models/model-name)
		id := strings.TrimPrefix(m.Name, "models/")
		models[i] = ModelInfo{
			ID:      id,
			Object:  "model",
			OwnedBy: "google",
		}
	}

	return &ModelsResponse{
		Object: "list",
		Data:   models,
	}, nil
}

func (p *GoogleProvider) Close() error {
	return nil
}

// toGeminiRequest converts internal request to Gemini format
func (p *GoogleProvider) toGeminiRequest(req *ChatRequest) map[string]interface{} {
	geminiReq := map[string]interface{}{}

	// Convert messages to Gemini format
	contents := make([]map[string]interface{}, 0)
	var systemInstruction string

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			if len(msg.Content) > 0 && msg.Content[0].Type == "text" {
				systemInstruction = msg.Content[0].Text
			}
			continue
		}

		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		parts := make([]map[string]interface{}, 0)
		for _, c := range msg.Content {
			switch c.Type {
			case "text":
				parts = append(parts, map[string]interface{}{
					"text": c.Text,
				})
			case "image_url":
				if c.ImageURL != nil {
					parts = append(parts, map[string]interface{}{
						"inlineData": map[string]interface{}{
							"mimeType": "image/jpeg",
							"data":     c.ImageURL.URL,
						},
					})
				}
			}
		}

		contents = append(contents, map[string]interface{}{
			"role":  role,
			"parts": parts,
		})
	}

	geminiReq["contents"] = contents

	if systemInstruction != "" {
		geminiReq["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": systemInstruction},
			},
		}
	}

	// Generation config
	genConfig := map[string]interface{}{}
	if req.Temperature != nil {
		genConfig["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		genConfig["topP"] = *req.TopP
	}
	if req.MaxTokens != nil {
		genConfig["maxOutputTokens"] = *req.MaxTokens
	}
	if len(req.Stop) > 0 {
		genConfig["stopSequences"] = req.Stop
	}

	if len(genConfig) > 0 {
		geminiReq["generationConfig"] = genConfig
	}

	return geminiReq
}

// geminiResponse represents Gemini's response format
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason  string `json:"finishReason"`
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings"`
	} `json:"candidates"`
	UsageMetadata *struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// toOpenAIResponse converts Gemini response to OpenAI format
func (p *GoogleProvider) toOpenAIResponse(resp *geminiResponse, model string) *ChatResponse {
	content := ""
	finishReason := "stop"

	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			content += part.Text
		}
		finishReason = p.mapFinishReason(resp.Candidates[0].FinishReason)
	}

	chatResp := &ChatResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []Choice{
			{
				Index: 0,
				Message: &ChatMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: finishReason,
			},
		},
	}

	if resp.UsageMetadata != nil {
		chatResp.Usage = &Usage{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		}
	}

	return chatResp
}

// toOpenAIStreamResponse converts Gemini stream response to OpenAI format
func (p *GoogleProvider) toOpenAIStreamResponse(resp *geminiResponse, model string) *ChatResponse {
	content := ""
	if len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			content += part.Text
		}
	}

	return &ChatResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []Choice{
			{
				Index: 0,
				Delta: &ChatMessage{
					Content: content,
				},
			},
		},
	}
}

func (p *GoogleProvider) mapFinishReason(reason string) string {
	switch reason {
	case "STOP":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	case "SAFETY":
		return "content_filter"
	default:
		return reason
	}
}

func init() {
	Register("google", NewGoogleProvider)
}
