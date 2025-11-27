# Configuration Reference

AINAR uses a YAML configuration file for all settings. By default, it looks for `config.yaml` in the current directory, or `~/.config/ainar/config.yaml`.

## Full Configuration Example

```yaml
# Server settings
server:
  # Port for the OpenAI-compatible API
  port: 8080
  # Host to bind to (use 0.0.0.0 for all interfaces)
  host: 127.0.0.1

# WebUI settings
webui:
  # Status: "disabled", "local", or "external"
  # WARNING: "external" allows anyone to access the WebUI!
  status: local
  # Port for the WebUI
  port: 8081

# Upstream providers
providers:
  - name: OpenAI
    code: openai
    api_type: openai
    base_url: https://api.openai.com/v1
    api_key: sk-your-api-key
    enabled: true

  - name: Anthropic
    code: anthropic
    api_type: anthropic
    base_url: https://api.anthropic.com
    api_key: sk-ant-your-key
    enabled: true

  - name: OpenRouter
    code: openrouter
    api_type: openai
    base_url: https://openrouter.ai/api/v1
    api_key: sk-or-your-key
    enabled: true

  - name: Local Ollama
    code: ollama
    api_type: ollama
    base_url: http://localhost:11434
    api_key: ""
    enabled: true

# Downstream API keys
downstream_keys:
  - name: my-app
    key: ainar-abc123...
    enabled: true
    limit:
      type: monthly      # daily, weekly, monthly, or total
      max_tokens: 1000000
    usage:
      tokens_used: 50000
      last_reset: 2024-01-01T00:00:00Z
    created_at: 2024-01-01T00:00:00Z

# Model aliases
aliases:
  - from: gpt4
    to: openai/gpt-4
  - from: claude
    to: anthropic/claude-3-sonnet-20240229
  - from: ora
    to: openrouter/anthropic
```

## Configuration Options

### Server Section

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `port` | integer | 8080 | Port for the API server |
| `host` | string | 127.0.0.1 | Host to bind to |

### WebUI Section

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `status` | string | local | One of: `disabled`, `local`, `external` |
| `port` | integer | 8081 | Port for the WebUI server |

**WebUI Status Values:**
- `disabled`: WebUI is completely disabled
- `local`: WebUI only accepts connections from localhost
- `external`: WebUI accepts connections from any IP (⚠️ Dangerous!)

### Provider Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `name` | string | Yes | Display name for the provider |
| `code` | string | Yes | Unique identifier (lowercase, numbers, `.`, `-`, `_`) |
| `api_type` | string | Yes | API type: `openai`, `anthropic`, `google`, `ollama` |
| `base_url` | string | Yes | Base URL for the API |
| `api_key` | string | No | API key (some providers don't require one) |
| `enabled` | boolean | No | Whether the provider is enabled (default: true) |
| `oauth` | object | No | OAuth configuration for providers like Copilot |

### Downstream Key Options

| Option | Type | Description |
|--------|------|-------------|
| `name` | string | Display name for the key |
| `key` | string | The actual API key |
| `enabled` | boolean | Whether the key is active |
| `limit` | object | Token limit configuration |
| `usage` | object | Current usage tracking |
| `created_at` | datetime | When the key was created |

### Token Limit Options

| Option | Type | Description |
|--------|------|-------------|
| `type` | string | `daily`, `weekly`, `monthly`, or `total` |
| `max_tokens` | integer | Maximum tokens allowed |

### Alias Options

| Option | Type | Description |
|--------|------|-------------|
| `from` | string | The alias to use |
| `to` | string | The full model path |

## Environment Variables

You can override some settings with environment variables:

| Variable | Description |
|----------|-------------|
| `AINAR_CONFIG` | Path to configuration file |
| `AINAR_PORT` | Override server port |
| `AINAR_HOST` | Override server host |

## Command Line Overrides

```bash
# Use custom config file
ainar serve --config /path/to/config.yaml

# Override ports
ainar serve --port 9000 --webui-port 9001

# Override host
ainar serve --host 0.0.0.0
```

## Security Considerations

1. **Protect your config file**: It contains API keys. Set appropriate permissions:
   ```bash
   chmod 600 config.yaml
   ```

2. **Never set WebUI to external** unless you understand the risks. Anyone who can access the WebUI can:
   - View all provider API keys
   - Generate new downstream keys
   - Modify configuration

3. **Use downstream keys**: Don't share your provider API keys. Generate downstream keys for each application.

4. **Set token limits**: Protect against runaway usage by setting limits on downstream keys.
