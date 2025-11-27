# API Reference

AINAR provides an OpenAI-compatible API for all downstream applications.

## Base URL

Default: `http://localhost:8080`

## Authentication

All requests require an API key in the `Authorization` header:

```
Authorization: Bearer YOUR_AINAR_KEY
```

## Endpoints

### Chat Completions

Create a chat completion.

**POST** `/v1/chat/completions`

**Request Body:**

```json
{
  "model": "openai/gpt-4",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "temperature": 0.7,
  "max_tokens": 1000,
  "stream": false
}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `model` | string | Yes | Model name with provider prefix (e.g., `openai/gpt-4`) |
| `messages` | array | Yes | Array of message objects |
| `temperature` | number | No | Sampling temperature (0-2) |
| `top_p` | number | No | Nucleus sampling parameter |
| `max_tokens` | integer | No | Maximum tokens to generate |
| `stream` | boolean | No | Enable streaming responses |
| `stop` | array | No | Stop sequences |
| `presence_penalty` | number | No | Presence penalty (-2 to 2) |
| `frequency_penalty` | number | No | Frequency penalty (-2 to 2) |
| `user` | string | No | User identifier |
| `tools` | array | No | Tools/functions for tool calling |

**Message Object:**

```json
{
  "role": "user|assistant|system",
  "content": "Message content"
}
```

**Multimodal Content:**

```json
{
  "role": "user",
  "content": [
    {"type": "text", "text": "What's in this image?"},
    {"type": "image_url", "image_url": {"url": "data:image/jpeg;base64,..."}}
  ]
}
```

**Response:**

```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 9,
    "completion_tokens": 12,
    "total_tokens": 21
  }
}
```

**Streaming Response:**

When `stream: true`, the response is sent as Server-Sent Events:

```
data: {"id":"chatcmpl-123","object":"chat.completion.chunk","choices":[{"delta":{"content":"Hello"}}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","choices":[{"delta":{"content":"!"}}]}

data: [DONE]
```

**Example:**

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Embeddings

Create embeddings for text.

**POST** `/v1/embeddings`

**Request Body:**

```json
{
  "model": "openai/text-embedding-3-small",
  "input": ["Hello, world!"]
}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `model` | string | Yes | Embedding model with provider prefix |
| `input` | string/array | Yes | Text(s) to embed |
| `encoding_format` | string | No | Output format (float or base64) |

**Response:**

```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "embedding": [0.0023064255, -0.009327292, ...],
      "index": 0
    }
  ],
  "model": "text-embedding-3-small",
  "usage": {
    "prompt_tokens": 5,
    "total_tokens": 5
  }
}
```

### List Models

List available models from all enabled providers.

**GET** `/v1/models`

**Response:**

```json
{
  "object": "list",
  "data": [
    {"id": "openai/gpt-4", "object": "model", "owned_by": "openai"},
    {"id": "openai/gpt-3.5-turbo", "object": "model", "owned_by": "openai"},
    {"id": "anthropic/claude-3-sonnet", "object": "model", "owned_by": "anthropic"}
  ]
}
```

### Get Model

Get information about a specific model.

**GET** `/v1/models/{model}`

**Example:**

```bash
curl http://localhost:8080/v1/models/openai/gpt-4 \
  -H "Authorization: Bearer YOUR_KEY"
```

**Response:**

```json
{
  "id": "openai/gpt-4",
  "object": "model",
  "owned_by": "openai"
}
```

## Model Naming Convention

Models are addressed using the format: `provider_code/model_name`

**Examples:**

| Model | Provider | Actual Model |
|-------|----------|--------------|
| `openai/gpt-4` | OpenAI | gpt-4 |
| `anthropic/claude-3-sonnet-20240229` | Anthropic | claude-3-sonnet-20240229 |
| `openrouter/anthropic/claude-3` | OpenRouter | anthropic/claude-3 |
| `ollama/llama2` | Ollama | llama2 |

## Using Aliases

If you've configured aliases, you can use them instead of full model names:

```bash
# With alias: gpt4 -> openai/gpt-4
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_KEY" \
  -d '{"model": "gpt4", "messages": [...]}'

# With prefix alias: ora -> openrouter/anthropic
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_KEY" \
  -d '{"model": "ora/claude-3", "messages": [...]}'
```

## Error Responses

**401 Unauthorized:**
```json
{
  "error": {
    "message": "Invalid API key",
    "type": "invalid_api_key",
    "code": 401
  }
}
```

**400 Bad Request:**
```json
{
  "error": {
    "message": "Invalid model name format: gpt-4 (expected provider/model)",
    "type": "invalid_model",
    "code": 400
  }
}
```

**429 Too Many Requests:**
```json
{
  "error": {
    "message": "Token limit exceeded: 100000/100000",
    "type": "rate_limit_exceeded",
    "code": 429
  }
}
```

**500 Internal Server Error:**
```json
{
  "error": {
    "message": "API error (status 401): Invalid API key",
    "type": "provider_error",
    "code": 500
  }
}
```

## Rate Limiting

Rate limiting is based on token usage per downstream key. Configure limits in the key settings:

- **Daily**: Resets every 24 hours
- **Weekly**: Resets every 7 days
- **Monthly**: Resets every 30 days
- **Total**: Never resets (hard limit)

When a limit is exceeded, requests return 429 Too Many Requests.
