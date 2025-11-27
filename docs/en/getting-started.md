# Getting Started

This guide will help you get AINAR up and running in a few minutes.

## Prerequisites

- A computer running Windows, macOS, or Linux
- At least one AI provider API key (OpenAI, Anthropic, etc.)

## Step 1: Install AINAR

### Option A: Download Binary

1. Go to the [Releases page](https://github.com/HeavySnowJakarta/ainar/releases)
2. Download the appropriate package for your system:
   - Windows: `ainar-windows-amd64.zip` or `ainar-windows-arm64.zip`
   - macOS: `ainar-macos-amd64.dmg` or `ainar-macos-arm64.dmg`
   - Linux: `ainar_VERSION_amd64.deb`, `ainar-VERSION-1.amd64.rpm`, or direct binary
3. Install/extract the package

### Option B: Build from Source

```bash
# Clone the repository
git clone https://github.com/HeavySnowJakarta/ainar.git
cd ainar

# Install Node.js dependencies and build frontend
cd web
npm install
npm run build
cd ..

# Copy frontend assets
cp -r web/dist/* internal/webui/dist/

# Build the Go binary
go build -o ainar ./cmd/ainar
```

## Step 2: Initialize Configuration

Create a default configuration file:

```bash
ainar init
```

This creates a `config.yaml` file with default settings.

## Step 3: Add a Provider

Add your first AI provider. For example, to add OpenAI:

```bash
ainar provider add \
  --name "OpenAI" \
  --code openai \
  --type openai \
  --url https://api.openai.com/v1 \
  --key sk-your-api-key
```

You can also add providers through the WebUI (see Step 5).

## Step 4: Generate a Downstream API Key

Generate an API key for your applications:

```bash
ainar key generate my-app
```

Save the generated key - it will only be shown once!

## Step 5: Start the Server

Start AINAR:

```bash
ainar serve
```

You should see:
```
API server listening on 127.0.0.1:8080
WebUI server listening on 127.0.0.1:8081
```

## Step 6: Access the WebUI

Open your browser and go to `http://localhost:8081` to access the WebUI.

From here you can:
- Add and manage providers
- Generate API keys
- Create model aliases

## Step 7: Use the API

Now you can use AINAR from your applications:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_AINAR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/gpt-4",
    "messages": [
      {"role": "user", "content": "Hello, world!"}
    ]
  }'
```

## Next Steps

- [Learn about configuration options](configuration.md)
- [Explore CLI commands](cli.md)
- [Set up as a system service](installation.md#system-service)
