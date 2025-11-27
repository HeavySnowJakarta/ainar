# AINAR - AI Model Router

[![CI](https://github.com/HeavySnowJakarta/ainar/actions/workflows/ci.yml/badge.svg)](https://github.com/HeavySnowJakarta/ainar/actions/workflows/ci.yml)
[![Release](https://github.com/HeavySnowJakarta/ainar/actions/workflows/release.yml/badge.svg)](https://github.com/HeavySnowJakarta/ainar/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

**AINAR (Ainar Is Not an AI Router)** is a unified OpenAI-compatible API gateway that routes requests to multiple AI providers. Configure once, use everywhere.

## Features

- 🔄 **Unified API**: Single OpenAI-compatible endpoint for all your AI providers
- 🔌 **Multiple Providers**: Support for OpenAI, Anthropic, Google, Ollama, and 30+ providers
- 🎯 **Model Routing**: Route requests to different providers using model name prefixes (e.g., `openai/gpt-4`, `anthropic/claude-3`)
- 📝 **Aliases**: Create shortcuts for long model names
- 🔑 **API Key Management**: Generate and manage downstream API keys with usage limits
- 🌐 **WebUI**: Easy-to-use web interface for configuration
- 🌍 **Internationalization**: WebUI available in 16 languages
- ⚡ **Lightweight**: Low memory footprint (~30MB), suitable for Raspberry Pi and similar devices
- 🚀 **Cross-platform**: Windows, macOS, and Linux support

## Quick Start

### Installation

#### Download Binary

Download the latest release from the [Releases page](https://github.com/HeavySnowJakarta/ainar/releases).

#### Build from Source

```bash
# Clone the repository
git clone https://github.com/HeavySnowJakarta/ainar.git
cd ainar

# Build
./scripts/build.sh
```

### Basic Usage

1. **Initialize configuration**:
   ```bash
   ainar init
   ```

2. **Add a provider**:
   ```bash
   ainar provider add --name "OpenAI" --code openai --url https://api.openai.com/v1 --key YOUR_API_KEY
   ```

3. **Generate a downstream API key**:
   ```bash
   ainar key generate my-app
   ```

4. **Start the server**:
   ```bash
   ainar serve
   ```

5. **Use the API**:
   ```bash
   curl http://localhost:8080/v1/chat/completions \
     -H "Authorization: Bearer YOUR_AINAR_KEY" \
     -H "Content-Type: application/json" \
     -d '{
       "model": "openai/gpt-4",
       "messages": [{"role": "user", "content": "Hello!"}]
     }'
   ```

## Documentation

Full documentation is available in multiple languages:

| Language | Documentation |
|----------|---------------|
| 🇬🇧 English | [docs/en](docs/en/README.md) |
| 🇪🇸 Español | [docs/es](docs/es/README.md) |
| 🇫🇷 Français | [docs/fr](docs/fr/README.md) |
| 🇩🇪 Deutsch | [docs/de](docs/de/README.md) |
| 🇵🇹 Português | [docs/pt](docs/pt/README.md) |
| 🇷🇺 Русский | [docs/ru](docs/ru/README.md) |
| 🇨🇳 简体中文 | [docs/zh](docs/zh/README.md) |
| 🇯🇵 日本語 | [docs/ja](docs/ja/README.md) |
| 🇰🇷 한국어 | [docs/ko](docs/ko/README.md) |
| 🇸🇦 العربية | [docs/ar](docs/ar/README.md) |
| 🇮🇳 हिन्दी | [docs/hi](docs/hi/README.md) |
| 🇮🇹 Italiano | [docs/it](docs/it/README.md) |
| 🇳🇱 Nederlands | [docs/nl](docs/nl/README.md) |
| 🇵🇱 Polski | [docs/pl](docs/pl/README.md) |
| 🇹🇷 Türkçe | [docs/tr](docs/tr/README.md) |
| 🇻🇳 Tiếng Việt | [docs/vi](docs/vi/README.md) |

> **Note**: Documentation in languages other than English may be outdated.

## Supported Providers

AINAR supports 30+ AI providers out of the box:

### First-party APIs
- OpenAI
- Anthropic (Claude)
- Google (Gemini)
- Mistral AI
- Cohere
- AI21 Labs
- xAI (Grok)

### OpenAI-compatible APIs
- OpenRouter
- Together AI
- Groq
- Fireworks AI
- Perplexity
- DeepSeek
- DeepInfra
- And many more...

### Local Models
- Ollama
- LM Studio
- vLLM
- llama.cpp
- Jan

## Configuration

AINAR uses a YAML configuration file. See [Configuration Reference](docs/en/configuration.md) for details.

```yaml
server:
  port: 8080
  host: 127.0.0.1

webui:
  status: local  # disabled, local, or external
  port: 8081

providers:
  - name: OpenAI
    code: openai
    api_type: openai
    base_url: https://api.openai.com/v1
    api_key: sk-...
    enabled: true

aliases:
  - from: gpt4
    to: openai/gpt-4

downstream_keys:
  - name: my-app
    key: ainar-...
    enabled: true
```

## System Service

AINAR can be installed as a system service:

- **Windows**: Installed as a Windows Service
- **macOS**: Uses launchd
- **Linux**: Uses systemd (or SysVinit as fallback)

See the [Installation Guide](docs/en/installation.md) for details.

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) first.

## License

AINAR is licensed under the [Apache License 2.0](LICENSE).
