# WebUI Guide

AINAR includes a web-based user interface for easy configuration and management.

## Accessing the WebUI

By default, the WebUI is available at `http://localhost:8081` when the server is running.

```bash
ainar serve
# WebUI server listening on 127.0.0.1:8081
```

Open your browser and navigate to the URL to access the WebUI.

## WebUI Modes

The WebUI can operate in three modes, controlled by the `webui.status` setting:

### Local Mode (Default)

```yaml
webui:
  status: local
```

- Only accessible from localhost (127.0.0.1)
- Safe for personal use
- Recommended for most users

### Disabled Mode

```yaml
webui:
  status: disabled
```

- WebUI is completely disabled
- Use CLI for all configuration
- Most secure option

### External Mode

```yaml
webui:
  status: external
```

⚠️ **WARNING**: This mode allows anyone who can reach the server to access the WebUI. This includes:
- Viewing all provider API keys
- Generating new downstream keys
- Modifying any configuration

Only use this mode if you have other security measures in place (e.g., VPN, firewall).

## Features

### Providers Tab

Manage your upstream AI providers.

![Providers Tab](images/providers-tab.png)

**Features:**
- View all configured providers
- Add new providers with auto-complete for known providers
- Edit existing provider settings
- Delete providers
- Enable/disable providers

**Adding a Provider:**

1. Click "Add Provider"
2. Start typing a provider name to see suggestions
3. Fill in the required fields:
   - **Name**: Display name (e.g., "OpenAI")
   - **Code**: Unique identifier (e.g., "openai")
   - **API Type**: Select the API format
   - **Base URL**: The API endpoint
   - **API Key**: Your API key
4. Click "Save"

### API Keys Tab

Manage downstream API keys for your applications.

![Keys Tab](images/keys-tab.png)

**Features:**
- View all generated keys
- See usage statistics and limits
- Generate new keys
- Enable/disable keys
- Delete keys

**Generating a Key:**

1. Click "Generate Key"
2. Enter a name for the key (e.g., "My App")
3. Optionally set a token limit:
   - **Limit Type**: Daily, Weekly, Monthly, or Total
   - **Max Tokens**: Maximum tokens allowed
4. Click "Generate"
5. **Important**: Copy and save the generated key immediately. It will only be shown once!

**Understanding Usage:**

For keys with limits, you'll see:
- Current usage count
- Maximum allowed tokens
- Usage percentage as a progress bar

Usage resets automatically based on the limit type:
- **Daily**: Resets every 24 hours
- **Weekly**: Resets every 7 days
- **Monthly**: Resets every 30 days
- **Total**: Never resets

### Aliases Tab

Create shortcuts for long model names.

![Aliases Tab](images/aliases-tab.png)

**Features:**
- View all configured aliases
- Add new aliases
- Delete aliases

**How Aliases Work:**

Aliases let you use shorter names for long model paths:

| Alias | Expands To | Example Usage |
|-------|------------|---------------|
| `gpt4` | `openai/gpt-4` | Use `gpt4` instead of `openai/gpt-4` |
| `ora` | `openrouter/anthropic` | Use `ora/claude-3` instead of `openrouter/anthropic/claude-3` |

**Adding an Alias:**

1. Click "Add Alias"
2. Enter the alias (e.g., "gpt4")
3. Enter the full model path (e.g., "openai/gpt-4")
4. Click "Save"

## Language Selection

The WebUI supports 16 languages. The language is automatically detected from your browser settings.

To change the language manually:
1. Look for the language selector in the top-right corner
2. Select your preferred language

**Supported Languages:**
- English
- Español (Spanish)
- Français (French)
- Deutsch (German)
- Português (Portuguese)
- Русский (Russian)
- 简体中文 (Simplified Chinese)
- 日本語 (Japanese)
- 한국어 (Korean)
- العربية (Arabic)
- हिन्दी (Hindi)
- Italiano (Italian)
- Nederlands (Dutch)
- Polski (Polish)
- Türkçe (Turkish)
- Tiếng Việt (Vietnamese)

## Accessibility

The WebUI is designed with accessibility in mind:

- Keyboard navigation support
- Screen reader friendly
- High contrast mode support
- Responsive design for mobile devices

## Troubleshooting

### Can't Access WebUI

1. **Check if server is running:**
   ```bash
   ainar serve
   ```

2. **Check WebUI status:**
   Make sure `webui.status` is not set to `disabled`

3. **Check the port:**
   Default is 8081. Verify it's not blocked by a firewall.

4. **Check if accessing from localhost:**
   In `local` mode, you can only access from 127.0.0.1

### WebUI Shows "Not Built"

If you see a message about the WebUI not being built:

1. Navigate to the `web` directory
2. Run `npm install`
3. Run `npm run build`
4. Copy the built files: `cp -r dist/* ../internal/webui/dist/`
5. Rebuild AINAR: `go build -o ainar ./cmd/ainar`
