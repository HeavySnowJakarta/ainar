// Package webui provides the web UI for AINAR
package webui

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/HeavySnowJakarta/ainar/internal/config"
	"github.com/HeavySnowJakarta/ainar/internal/providers"
)

//go:embed dist/*
var staticFiles embed.FS

// Handler provides the WebUI HTTP handler
type Handler struct {
	config     *config.Config
	configPath string
	localOnly  bool
}

// NewHandler creates a new WebUI handler
func NewHandler(cfg *config.Config, configPath string, localOnly bool) *Handler {
	return &Handler{
		config:     cfg,
		configPath: configPath,
		localOnly:  localOnly,
	}
}

// ServeHTTP handles HTTP requests for the WebUI
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if local only
	if h.localOnly {
		remoteAddr := r.RemoteAddr
		if !strings.HasPrefix(remoteAddr, "127.0.0.1:") &&
			!strings.HasPrefix(remoteAddr, "[::1]:") &&
			!strings.HasPrefix(remoteAddr, "localhost:") {
			http.Error(w, "Forbidden: WebUI is set to local only", http.StatusForbidden)
			return
		}
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// API routes
	if strings.HasPrefix(r.URL.Path, "/api/") {
		h.handleAPI(w, r)
		return
	}

	// Serve static files
	h.serveStatic(w, r)
}

// handleAPI handles API requests
func (h *Handler) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")

	switch {
	// Providers
	case path == "providers" && r.Method == "GET":
		h.listProviders(w, r)
	case path == "providers" && r.Method == "POST":
		h.addProvider(w, r)
	case strings.HasPrefix(path, "providers/") && r.Method == "PUT":
		h.updateProvider(w, r, strings.TrimPrefix(path, "providers/"))
	case strings.HasPrefix(path, "providers/") && r.Method == "DELETE":
		h.deleteProvider(w, r, strings.TrimPrefix(path, "providers/"))

	// Downstream Keys
	case path == "keys" && r.Method == "GET":
		h.listKeys(w, r)
	case path == "keys" && r.Method == "POST":
		h.createKey(w, r)
	case strings.HasPrefix(path, "keys/") && r.Method == "PUT":
		h.updateKey(w, r, strings.TrimPrefix(path, "keys/"))
	case strings.HasPrefix(path, "keys/") && r.Method == "DELETE":
		h.deleteKey(w, r, strings.TrimPrefix(path, "keys/"))

	// Aliases
	case path == "aliases" && r.Method == "GET":
		h.listAliases(w, r)
	case path == "aliases" && r.Method == "POST":
		h.addAlias(w, r)
	case strings.HasPrefix(path, "aliases/") && r.Method == "DELETE":
		h.deleteAlias(w, r, strings.TrimPrefix(path, "aliases/"))

	// Known providers
	case path == "known-providers" && r.Method == "GET":
		h.listKnownProviders(w, r)

	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

// Provider handlers
func (h *Handler) listProviders(w http.ResponseWriter, r *http.Request) {
	// Return providers without API keys for security
	safeProviders := make([]map[string]interface{}, len(h.config.Providers))
	for i, p := range h.config.Providers {
		safeProviders[i] = map[string]interface{}{
			"name":     p.Name,
			"code":     p.Code,
			"api_type": p.APIType,
			"base_url": p.BaseURL,
			"enabled":  p.Enabled,
			"has_key":  p.APIKey != "",
		}
	}
	writeJSON(w, safeProviders)
}

func (h *Handler) addProvider(w http.ResponseWriter, r *http.Request) {
	var p config.Provider
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	p.Enabled = true
	if err := h.config.AddProvider(p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.config.Save(h.configPath); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *Handler) updateProvider(w http.ResponseWriter, r *http.Request, code string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	provider, err := h.config.GetProvider(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		provider.Name = name
	}
	if baseURL, ok := updates["base_url"].(string); ok {
		provider.BaseURL = baseURL
	}
	if apiKey, ok := updates["api_key"].(string); ok {
		provider.APIKey = apiKey
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		provider.Enabled = enabled
	}

	if err := h.config.Save(h.configPath); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *Handler) deleteProvider(w http.ResponseWriter, r *http.Request, code string) {
	if err := h.config.RemoveProvider(code); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := h.config.Save(h.configPath); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

// Key handlers
func (h *Handler) listKeys(w http.ResponseWriter, r *http.Request) {
	// Return keys without actual key values
	safeKeys := make([]map[string]interface{}, len(h.config.DownstreamKeys))
	for i, k := range h.config.DownstreamKeys {
		keyInfo := map[string]interface{}{
			"name":       k.Name,
			"created_at": k.CreatedAt,
			"enabled":    k.Enabled,
		}
		if k.Limit != nil {
			keyInfo["limit"] = k.Limit
		}
		if k.Usage != nil {
			keyInfo["usage"] = k.Usage
		}
		safeKeys[i] = keyInfo
	}
	writeJSON(w, safeKeys)
}

func (h *Handler) createKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string             `json:"name"`
		Limit *config.TokenLimit `json:"limit,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	key, err := h.config.GenerateDownstreamKey(req.Name, req.Limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.config.Save(h.configPath); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"key": key})
}

func (h *Handler) updateKey(w http.ResponseWriter, r *http.Request, name string) {
	var updates struct {
		Enabled *bool              `json:"enabled,omitempty"`
		Limit   *config.TokenLimit `json:"limit,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find and update key
	found := false
	for i := range h.config.DownstreamKeys {
		if h.config.DownstreamKeys[i].Name == name {
			if updates.Enabled != nil {
				h.config.DownstreamKeys[i].Enabled = *updates.Enabled
			}
			if updates.Limit != nil {
				h.config.DownstreamKeys[i].Limit = updates.Limit
				// Reset usage when limit is changed
				h.config.DownstreamKeys[i].Usage = &config.TokenUsage{}
			}
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	if err := h.config.Save(h.configPath); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *Handler) deleteKey(w http.ResponseWriter, r *http.Request, name string) {
	for i := range h.config.DownstreamKeys {
		if h.config.DownstreamKeys[i].Name == name {
			h.config.DownstreamKeys = append(h.config.DownstreamKeys[:i], h.config.DownstreamKeys[i+1:]...)
			break
		}
	}

	if err := h.config.Save(h.configPath); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

// Alias handlers
func (h *Handler) listAliases(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.config.Aliases)
}

func (h *Handler) addAlias(w http.ResponseWriter, r *http.Request) {
	var alias config.Alias
	if err := json.NewDecoder(r.Body).Decode(&alias); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.config.AddAlias(alias.From, alias.To); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.config.Save(h.configPath); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *Handler) deleteAlias(w http.ResponseWriter, r *http.Request, from string) {
	for i := range h.config.Aliases {
		if h.config.Aliases[i].From == from {
			h.config.Aliases = append(h.config.Aliases[:i], h.config.Aliases[i+1:]...)
			break
		}
	}

	if err := h.config.Save(h.configPath); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

// Known providers handler
func (h *Handler) listKnownProviders(w http.ResponseWriter, r *http.Request) {
	apiType := r.URL.Query().Get("api_type")
	query := r.URL.Query().Get("q")

	var knownProviders []providers.KnownProvider

	if apiType != "" {
		knownProviders = providers.FindKnownProviderByAPIType(apiType)
	} else if query != "" {
		knownProviders = providers.FindKnownProvider(query)
	} else {
		knownProviders = providers.KnownProviders()
	}

	writeJSON(w, knownProviders)
}

// serveStatic serves static files from the embedded filesystem
func (h *Handler) serveStatic(w http.ResponseWriter, r *http.Request) {
	// Try to get the embedded filesystem
	subFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		// If dist doesn't exist, serve a placeholder
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AINAR - WebUI</title>
    <style>
        body { font-family: system-ui, sans-serif; padding: 2rem; max-width: 800px; margin: 0 auto; }
        h1 { color: #333; }
        .warning { background: #fff3cd; border: 1px solid #ffc107; padding: 1rem; border-radius: 0.5rem; margin: 1rem 0; }
    </style>
</head>
<body>
    <h1>AINAR WebUI</h1>
    <div class="warning">
        <strong>Note:</strong> The WebUI frontend is not built yet. 
        Please run <code>npm run build</code> in the <code>web</code> directory.
    </div>
    <h2>API Endpoints</h2>
    <ul>
        <li><strong>GET /api/providers</strong> - List providers</li>
        <li><strong>POST /api/providers</strong> - Add provider</li>
        <li><strong>GET /api/keys</strong> - List downstream keys</li>
        <li><strong>POST /api/keys</strong> - Create downstream key</li>
        <li><strong>GET /api/aliases</strong> - List aliases</li>
        <li><strong>POST /api/aliases</strong> - Add alias</li>
        <li><strong>GET /api/known-providers</strong> - List known providers</li>
    </ul>
</body>
</html>`))
		return
	}

	// Serve files
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// Try to open the file
	file, err := subFS.Open(strings.TrimPrefix(path, "/"))
	if err != nil {
		// Serve index.html for SPA routing
		file, err = subFS.Open("index.html")
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	// Set content type
	contentType := getContentType(path)
	w.Header().Set("Content-Type", contentType)

	// Serve file
	if seeker, ok := file.(fs.File); ok {
		http.ServeContent(w, r, info.Name(), info.ModTime(), seeker.(interface {
			fs.File
			Seek(int64, int) (int64, error)
		}))
	}
}

func getContentType(path string) string {
	switch {
	case strings.HasSuffix(path, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(path, ".json"):
		return "application/json"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".ico"):
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
