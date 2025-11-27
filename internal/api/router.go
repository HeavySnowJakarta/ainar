// Package api provides the OpenAI-compatible API router
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/HeavySnowJakarta/ainar/internal/config"
	"github.com/HeavySnowJakarta/ainar/internal/providers"
)

// Router handles incoming API requests
type Router struct {
	config         *config.Config
	providerCache  map[string]providers.Provider
	providerMutex  sync.RWMutex
	configPath     string
}

// NewRouter creates a new API router
func NewRouter(cfg *config.Config, configPath string) *Router {
	return &Router{
		config:        cfg,
		providerCache: make(map[string]providers.Provider),
		configPath:    configPath,
	}
}

// ServeHTTP handles incoming HTTP requests
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Validate API key
	apiKey := extractAPIKey(req)
	dk, err := r.config.ValidateDownstreamKey(apiKey)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_api_key", "Invalid API key")
		return
	}

	// Check token limits
	if dk.Limit != nil && dk.Usage != nil {
		if err := r.checkTokenLimit(dk); err != nil {
			writeError(w, http.StatusTooManyRequests, "rate_limit_exceeded", err.Error())
			return
		}
	}

	// Route request
	path := req.URL.Path
	switch {
	case path == "/v1/chat/completions" && req.Method == "POST":
		r.handleChatCompletions(w, req, dk)
	case path == "/v1/embeddings" && req.Method == "POST":
		r.handleEmbeddings(w, req, dk)
	case path == "/v1/models" && req.Method == "GET":
		r.handleListModels(w, req)
	case strings.HasPrefix(path, "/v1/models/") && req.Method == "GET":
		r.handleGetModel(w, req)
	default:
		writeError(w, http.StatusNotFound, "not_found", "Endpoint not found")
	}
}

// handleChatCompletions handles chat completion requests
func (r *Router) handleChatCompletions(w http.ResponseWriter, req *http.Request, dk *config.DownstreamKey) {
	var chatReq providers.ChatRequest
	if err := json.NewDecoder(req.Body).Decode(&chatReq); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Resolve alias
	model := r.config.ResolveAlias(chatReq.Model)

	// Parse provider and model from model name
	providerCode, modelName, err := parseModelName(model)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_model", err.Error())
		return
	}

	// Get provider
	provider, err := r.getProvider(providerCode)
	if err != nil {
		writeError(w, http.StatusBadRequest, "provider_error", err.Error())
		return
	}

	// Update model name to just the model part
	chatReq.Model = modelName

	if chatReq.Stream {
		r.handleStreamingChat(w, req, provider, &chatReq, dk)
	} else {
		r.handleNonStreamingChat(w, req, provider, &chatReq, dk)
	}
}

// handleNonStreamingChat handles non-streaming chat requests
func (r *Router) handleNonStreamingChat(w http.ResponseWriter, req *http.Request, provider providers.Provider, chatReq *providers.ChatRequest, dk *config.DownstreamKey) {
	resp, err := provider.Chat(req.Context(), chatReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "provider_error", err.Error())
		return
	}

	// Update token usage
	if resp.Usage != nil && dk != nil {
		r.updateTokenUsage(dk, resp.Usage.TotalTokens)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeError(w, http.StatusInternalServerError, "encoding_error", "Failed to encode response")
	}
}

// handleStreamingChat handles streaming chat requests
func (r *Router) handleStreamingChat(w http.ResponseWriter, req *http.Request, provider providers.Provider, chatReq *providers.ChatRequest, dk *config.DownstreamKey) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming_error", "Streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	respChan, errChan := provider.ChatStream(req.Context(), chatReq)

	totalTokens := 0
	for {
		select {
		case resp, ok := <-respChan:
			if !ok {
				// Stream ended
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()

				// Update token usage
				if totalTokens > 0 && dk != nil {
					r.updateTokenUsage(dk, totalTokens)
				}
				return
			}

			// Estimate tokens from content
			if len(resp.Choices) > 0 && resp.Choices[0].Delta != nil {
				totalTokens += len(resp.Choices[0].Delta.Content) / 4 // rough estimate
			}

			data, err := json.Marshal(resp)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

		case err := <-errChan:
			if err != nil {
				errResp := map[string]interface{}{
					"error": map[string]interface{}{
						"message": err.Error(),
						"type":    "provider_error",
					},
				}
				data, _ := json.Marshal(errResp)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
			return

		case <-req.Context().Done():
			return
		}
	}
}

// handleEmbeddings handles embedding requests
func (r *Router) handleEmbeddings(w http.ResponseWriter, req *http.Request, dk *config.DownstreamKey) {
	var embReq providers.EmbeddingRequest
	body, err := io.ReadAll(req.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Failed to read request body")
		return
	}

	if err := json.Unmarshal(body, &embReq); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Handle different input types (string or array)
	var rawReq map[string]interface{}
	if err := json.Unmarshal(body, &rawReq); err == nil {
		if input, ok := rawReq["input"]; ok {
			switch v := input.(type) {
			case string:
				embReq.Input = []string{v}
			case []interface{}:
				embReq.Input = make([]string, len(v))
				for i, item := range v {
					if s, ok := item.(string); ok {
						embReq.Input[i] = s
					}
				}
			}
		}
	}

	// Resolve alias
	model := r.config.ResolveAlias(embReq.Model)

	// Parse provider and model
	providerCode, modelName, err := parseModelName(model)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_model", err.Error())
		return
	}

	provider, err := r.getProvider(providerCode)
	if err != nil {
		writeError(w, http.StatusBadRequest, "provider_error", err.Error())
		return
	}

	embReq.Model = modelName

	resp, err := provider.Embedding(req.Context(), &embReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "provider_error", err.Error())
		return
	}

	// Update token usage
	if resp.Usage != nil && dk != nil {
		r.updateTokenUsage(dk, resp.Usage.TotalTokens)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeError(w, http.StatusInternalServerError, "encoding_error", "Failed to encode response")
	}
}

// handleListModels handles model listing requests
func (r *Router) handleListModels(w http.ResponseWriter, req *http.Request) {
	var allModels []providers.ModelInfo

	for _, p := range r.config.Providers {
		if !p.Enabled {
			continue
		}

		provider, err := r.getProvider(p.Code)
		if err != nil {
			continue
		}

		models, err := provider.ListModels(req.Context())
		if err != nil {
			continue
		}

		// Prefix model names with provider code
		for _, m := range models.Data {
			m.ID = p.Code + "/" + m.ID
			allModels = append(allModels, m)
		}
	}

	resp := providers.ModelsResponse{
		Object: "list",
		Data:   allModels,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleGetModel handles single model retrieval
func (r *Router) handleGetModel(w http.ResponseWriter, req *http.Request) {
	modelID := strings.TrimPrefix(req.URL.Path, "/v1/models/")

	providerCode, modelName, err := parseModelName(modelID)
	if err != nil {
		writeError(w, http.StatusNotFound, "model_not_found", "Model not found")
		return
	}

	provider, err := r.getProvider(providerCode)
	if err != nil {
		writeError(w, http.StatusNotFound, "model_not_found", "Provider not found")
		return
	}

	models, err := provider.ListModels(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "provider_error", err.Error())
		return
	}

	for _, m := range models.Data {
		if m.ID == modelName {
			m.ID = modelID
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(m)
			return
		}
	}

	writeError(w, http.StatusNotFound, "model_not_found", "Model not found")
}

// getProvider returns a cached or new provider instance
func (r *Router) getProvider(code string) (providers.Provider, error) {
	r.providerMutex.RLock()
	if provider, ok := r.providerCache[code]; ok {
		r.providerMutex.RUnlock()
		return provider, nil
	}
	r.providerMutex.RUnlock()

	r.providerMutex.Lock()
	defer r.providerMutex.Unlock()

	// Double-check after acquiring write lock
	if provider, ok := r.providerCache[code]; ok {
		return provider, nil
	}

	// Find provider config
	providerCfg, err := r.config.GetProvider(code)
	if err != nil {
		return nil, err
	}

	if !providerCfg.Enabled {
		return nil, fmt.Errorf("provider %s is disabled", code)
	}

	// Create provider
	provider, err := providers.Create(providerCfg.APIType, providerCfg.BaseURL, providerCfg.APIKey)
	if err != nil {
		return nil, err
	}

	r.providerCache[code] = provider
	return provider, nil
}

// checkTokenLimit checks if the token limit has been exceeded
func (r *Router) checkTokenLimit(dk *config.DownstreamKey) error {
	if dk.Limit == nil || dk.Usage == nil {
		return nil
	}

	// Check if we need to reset the usage
	now := time.Now()
	shouldReset := false

	switch dk.Limit.Type {
	case "daily":
		shouldReset = now.Sub(dk.Usage.LastReset) >= 24*time.Hour
	case "weekly":
		shouldReset = now.Sub(dk.Usage.LastReset) >= 7*24*time.Hour
	case "monthly":
		shouldReset = now.Sub(dk.Usage.LastReset) >= 30*24*time.Hour
	}

	if shouldReset {
		dk.Usage.TokensUsed = 0
		dk.Usage.LastReset = now
		r.config.Save(r.configPath)
	}

	if dk.Usage.TokensUsed >= dk.Limit.MaxTokens {
		return fmt.Errorf("token limit exceeded: %d/%d", dk.Usage.TokensUsed, dk.Limit.MaxTokens)
	}

	return nil
}

// updateTokenUsage updates the token usage for a downstream key
func (r *Router) updateTokenUsage(dk *config.DownstreamKey, tokens int) {
	if dk.Usage == nil {
		dk.Usage = &config.TokenUsage{LastReset: time.Now()}
	}
	dk.Usage.TokensUsed += int64(tokens)
	r.config.Save(r.configPath)
}

// ClearProviderCache clears the provider cache
func (r *Router) ClearProviderCache() {
	r.providerMutex.Lock()
	defer r.providerMutex.Unlock()

	for _, p := range r.providerCache {
		p.Close()
	}
	r.providerCache = make(map[string]providers.Provider)
}

// parseModelName parses a model name into provider code and model name
func parseModelName(model string) (providerCode, modelName string, err error) {
	parts := strings.SplitN(model, "/", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid model name format: %s (expected provider/model)", model)
	}
	return parts[0], parts[1], nil
}

// extractAPIKey extracts the API key from the request
func extractAPIKey(req *http.Request) string {
	auth := req.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, errType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    errType,
			"code":    status,
		},
	})
}
