# CLI Reference

AINAR provides a command-line interface for all operations.

## Global Options

```
--config, -c    Path to configuration file
--help, -h      Show help
```

## Commands

### ainar

Run without arguments to start the server with default settings.

```bash
ainar
```

### ainar version

Display version information.

```bash
ainar version
```

Output:
```
AINAR version v1.0.0
Git commit: abc1234
Build date: 2024-01-01T00:00:00Z
```

### ainar init

Initialize a new configuration file.

```bash
ainar init
```

Creates a default `config.yaml` in the current directory (or XDG config directory if not writable).

### ainar serve

Start the AINAR server.

```bash
ainar serve [flags]
```

**Flags:**
```
--host          Host to bind to (overrides config)
--port          API port (overrides config)
--webui-port    WebUI port (overrides config)
```

**Examples:**
```bash
# Start with default settings
ainar serve

# Bind to all interfaces on custom ports
ainar serve --host 0.0.0.0 --port 9000 --webui-port 9001

# Use specific config file
ainar serve --config /etc/ainar/config.yaml
```

### ainar provider

Manage upstream providers.

#### ainar provider list

List all configured providers.

```bash
ainar provider list
```

Output:
```
openai          OpenAI               openai     https://api.openai.com/v1 [enabled]
anthropic       Anthropic            anthropic  https://api.anthropic.com [enabled]
ollama          Local Ollama         ollama     http://localhost:11434 [disabled]
```

#### ainar provider add

Add a new provider.

```bash
ainar provider add [flags]
```

**Flags:**
```
--name, -n      Provider name (required)
--code, -c      Provider code (required)
--type, -t      API type: openai, anthropic, google, ollama (default: openai)
--url, -u       Base URL (required)
--key, -k       API key
```

**Examples:**
```bash
# Add OpenAI
ainar provider add -n "OpenAI" -c openai -t openai -u https://api.openai.com/v1 -k sk-your-key

# Add Anthropic
ainar provider add -n "Anthropic" -c anthropic -t anthropic -u https://api.anthropic.com -k sk-ant-your-key

# Add local Ollama (no API key needed)
ainar provider add -n "Ollama" -c ollama -t ollama -u http://localhost:11434

# Add OpenRouter (OpenAI-compatible)
ainar provider add -n "OpenRouter" -c openrouter -t openai -u https://openrouter.ai/api/v1 -k sk-or-your-key
```

#### ainar provider remove

Remove a provider.

```bash
ainar provider remove <code>
```

**Example:**
```bash
ainar provider remove openai
```

### ainar key

Manage downstream API keys.

#### ainar key list

List all downstream keys.

```bash
ainar key list
```

Output:
```
my-app              [enabled] monthly: 1000000 tokens (5.0% used)
test-key            [disabled] no limit
```

#### ainar key generate

Generate a new downstream key.

```bash
ainar key generate <name> [flags]
```

**Flags:**
```
--limit-type    Limit type: daily, weekly, monthly, total
--limit-tokens  Maximum tokens
```

**Examples:**
```bash
# Generate key without limits
ainar key generate my-app

# Generate key with monthly limit
ainar key generate prod-app --limit-type monthly --limit-tokens 1000000

# Generate key with daily limit
ainar key generate dev-app --limit-type daily --limit-tokens 10000
```

Output:
```
API key generated for my-app:

  ainar-abc123def456...

⚠️  Save this key now. It will only be shown once!
```

#### ainar key delete

Delete a downstream key.

```bash
ainar key delete <name>
```

**Example:**
```bash
ainar key delete test-key
```

### ainar alias

Manage model aliases.

#### ainar alias list

List all aliases.

```bash
ainar alias list
```

Output:
```
gpt4 -> openai/gpt-4
claude -> anthropic/claude-3-sonnet-20240229
ora -> openrouter/anthropic
```

#### ainar alias add

Add a new alias.

```bash
ainar alias add <from> <to>
```

**Examples:**
```bash
# Simple alias
ainar alias add gpt4 openai/gpt-4

# Alias for a prefix
ainar alias add ora openrouter/anthropic

# Then use as: ora/claude-3-sonnet-20240229
```

#### ainar alias remove

Remove an alias.

```bash
ainar alias remove <from>
```

**Example:**
```bash
ainar alias remove gpt4
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |

## Configuration File Location

AINAR looks for configuration in this order:

1. Path specified with `--config`
2. `config.yaml` in current directory
3. `~/.config/ainar/config.yaml`

## Examples

### Complete Setup

```bash
# Initialize config
ainar init

# Add providers
ainar provider add -n "OpenAI" -c openai -u https://api.openai.com/v1 -k sk-...
ainar provider add -n "Anthropic" -c anthropic -t anthropic -u https://api.anthropic.com -k sk-ant-...

# Add aliases
ainar alias add gpt4 openai/gpt-4
ainar alias add claude anthropic/claude-3-sonnet-20240229

# Generate API key
ainar key generate my-app --limit-type monthly --limit-tokens 1000000

# Start server
ainar serve
```

### Using with systemd

```bash
# Start server in background (handled by systemd)
sudo systemctl start ainar

# Check status
sudo systemctl status ainar

# View logs
sudo journalctl -u ainar -f
```
