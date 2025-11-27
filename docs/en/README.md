# AINAR Documentation

Welcome to the AINAR documentation. AINAR (Ainar Is Not an AI Router) is a unified OpenAI-compatible API gateway that routes requests to multiple AI providers.

## Table of Contents

- [Getting Started](getting-started.md)
- [Installation](installation.md)
- [Configuration](configuration.md)
- [CLI Reference](cli.md)
- [WebUI Guide](webui.md)
- [API Reference](api.md)

## Quick Links

- [GitHub Repository](https://github.com/HeavySnowJakarta/ainar)
- [Releases](https://github.com/HeavySnowJakarta/ainar/releases)
- [Issue Tracker](https://github.com/HeavySnowJakarta/ainar/issues)

## Overview

AINAR provides a single OpenAI-compatible endpoint that can route requests to multiple AI providers based on model name prefixes. This allows you to:

1. **Configure once**: Set up all your AI providers in one place
2. **Use everywhere**: Use the same API endpoint for all your applications
3. **Simplify management**: Manage API keys and usage limits centrally

### How It Works

1. Your application sends a request to AINAR with a model name like `openai/gpt-4`
2. AINAR extracts the provider code (`openai`) and model name (`gpt-4`)
3. AINAR routes the request to the appropriate provider
4. The response is returned in OpenAI-compatible format

### Example

```bash
# Request to OpenAI
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_AINAR_KEY" \
  -d '{"model": "openai/gpt-4", "messages": [...]}'

# Request to Anthropic
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_AINAR_KEY" \
  -d '{"model": "anthropic/claude-3-sonnet", "messages": [...]}'
```

## Support

If you encounter any issues, please [open an issue](https://github.com/HeavySnowJakarta/ainar/issues/new) on GitHub.
