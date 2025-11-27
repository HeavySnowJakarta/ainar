package providers

// KnownProvider represents a well-known AI provider
type KnownProvider struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	APIType string `json:"api_type"`
	BaseURL string `json:"base_url"`
}

// KnownProviders returns a list of well-known AI providers
func KnownProviders() []KnownProvider {
	return []KnownProvider{
		// OpenAI and compatible
		{Name: "OpenAI", Code: "openai", APIType: "openai", BaseURL: "https://api.openai.com/v1"},
		{Name: "Azure OpenAI", Code: "azure", APIType: "openai", BaseURL: "https://{resource}.openai.azure.com/openai/deployments/{deployment}"},
		{Name: "OpenRouter", Code: "openrouter", APIType: "openai", BaseURL: "https://openrouter.ai/api/v1"},
		{Name: "Together AI", Code: "together", APIType: "openai", BaseURL: "https://api.together.xyz/v1"},
		{Name: "Fireworks AI", Code: "fireworks", APIType: "openai", BaseURL: "https://api.fireworks.ai/inference/v1"},
		{Name: "Groq", Code: "groq", APIType: "openai", BaseURL: "https://api.groq.com/openai/v1"},
		{Name: "Perplexity", Code: "perplexity", APIType: "openai", BaseURL: "https://api.perplexity.ai"},
		{Name: "Mistral AI", Code: "mistral", APIType: "openai", BaseURL: "https://api.mistral.ai/v1"},
		{Name: "DeepSeek", Code: "deepseek", APIType: "openai", BaseURL: "https://api.deepseek.com"},
		{Name: "Deepinfra", Code: "deepinfra", APIType: "openai", BaseURL: "https://api.deepinfra.com/v1/openai"},
		{Name: "Anyscale", Code: "anyscale", APIType: "openai", BaseURL: "https://api.endpoints.anyscale.com/v1"},
		{Name: "Lepton AI", Code: "lepton", APIType: "openai", BaseURL: "https://llm.lepton.ai/api/v1"},
		{Name: "Novita AI", Code: "novita", APIType: "openai", BaseURL: "https://api.novita.ai/v3/openai"},
		{Name: "SiliconFlow", Code: "siliconflow", APIType: "openai", BaseURL: "https://api.siliconflow.cn/v1"},
		{Name: "Moonshot AI", Code: "moonshot", APIType: "openai", BaseURL: "https://api.moonshot.cn/v1"},
		{Name: "ZhipuAI", Code: "zhipu", APIType: "openai", BaseURL: "https://open.bigmodel.cn/api/paas/v4"},
		{Name: "Yi AI", Code: "yi", APIType: "openai", BaseURL: "https://api.lingyiwanwu.com/v1"},
		{Name: "Baichuan AI", Code: "baichuan", APIType: "openai", BaseURL: "https://api.baichuan-ai.com/v1"},
		{Name: "Qwen (Alibaba)", Code: "qwen", APIType: "openai", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1"},
		{Name: "Cerebras", Code: "cerebras", APIType: "openai", BaseURL: "https://api.cerebras.ai/v1"},
		{Name: "LMStudio", Code: "lmstudio", APIType: "openai", BaseURL: "http://localhost:1234/v1"},
		{Name: "vLLM", Code: "vllm", APIType: "openai", BaseURL: "http://localhost:8000/v1"},
		{Name: "llama.cpp", Code: "llamacpp", APIType: "openai", BaseURL: "http://localhost:8080/v1"},
		{Name: "Jan", Code: "jan", APIType: "openai", BaseURL: "http://localhost:1337/v1"},
		{Name: "Hyperbolic", Code: "hyperbolic", APIType: "openai", BaseURL: "https://api.hyperbolic.xyz/v1"},
		{Name: "AI21 Labs", Code: "ai21", APIType: "openai", BaseURL: "https://api.ai21.com/studio/v1"},
		{Name: "Cohere", Code: "cohere", APIType: "openai", BaseURL: "https://api.cohere.ai/v1"},
		{Name: "xAI (Grok)", Code: "xai", APIType: "openai", BaseURL: "https://api.x.ai/v1"},

		// Anthropic
		{Name: "Anthropic", Code: "anthropic", APIType: "anthropic", BaseURL: "https://api.anthropic.com"},
		{Name: "Amazon Bedrock (Claude)", Code: "bedrock-claude", APIType: "anthropic", BaseURL: "https://bedrock-runtime.{region}.amazonaws.com"},

		// Google
		{Name: "Google AI (Gemini)", Code: "google", APIType: "google", BaseURL: "https://generativelanguage.googleapis.com"},
		{Name: "Google Vertex AI", Code: "vertex", APIType: "google", BaseURL: "https://{region}-aiplatform.googleapis.com"},

		// Ollama
		{Name: "Ollama", Code: "ollama", APIType: "ollama", BaseURL: "http://localhost:11434"},
	}
}

// FindKnownProvider finds a known provider by name or code
func FindKnownProvider(query string) []KnownProvider {
	query = normalizeQuery(query)
	var matches []KnownProvider

	for _, p := range KnownProviders() {
		if contains(p.Name, query) || contains(p.Code, query) {
			matches = append(matches, p)
		}
	}

	return matches
}

// FindKnownProviderByAPIType finds known providers by API type
func FindKnownProviderByAPIType(apiType string) []KnownProvider {
	var matches []KnownProvider

	for _, p := range KnownProviders() {
		if p.APIType == apiType {
			matches = append(matches, p)
		}
	}

	return matches
}

func normalizeQuery(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + 32 // lowercase
		}
		result = append(result, c)
	}
	return string(result)
}

func contains(s, substr string) bool {
	s = normalizeQuery(s)
	substr = normalizeQuery(substr)

	if len(substr) > len(s) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
