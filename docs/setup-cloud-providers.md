# Cloud Provider Setup Guide

This guide explains how to set up free cloud providers for LLM and embeddings when you don't want to run models locally.

## Table of Contents

1. [Groq (LLM)](#groq-llm)
2. [Hugging Face (Embeddings)](#hugging-face-embeddings)
3. [Configuration](#configuration)
4. [Testing](#testing)

---

## Groq (LLM)

Groq provides a free tier with generous limits for development:
- **30 requests per minute**
- **7,000 requests per day**
- **Fast inference** (LPU hardware)

### Step 1: Create Account

1. Go to [https://console.groq.com](https://console.groq.com)
2. Sign up with GitHub, Google, or email
3. Verify your email

### Step 2: Get API Key

1. Go to [https://console.groq.com/keys](https://console.groq.com/keys)
2. Click **"Create API Key"**
3. Give it a name (e.g., "ecommerce-ai-assistant")
4. Copy the key (starts with `gsk_...`)
5. **Store it securely** - you won't be able to see it again

### Step 3: Choose Model

Groq supports several models. For this project, we recommend:

| Model | Size | Speed | Quality | Best For |
|-------|------|-------|---------|----------|
| `llama-3.2-3b-preview` | 3B params | Very fast | Good | General purpose |
| `llama-3.1-8b-instant` | 8B params | Fast | Better | Complex reasoning |
| `gemma2-9b-it` | 9B params | Fast | Better | Instruction following |

**Recommendation:** Start with `llama-3.2-3b-preview` - it's fast and works well for tool calling.

### Step 4: Configure

```powershell
# Set environment variables
$env:LLM_BASE_URL="https://api.groq.com/openai/v1"
$env:LLM_API_KEY="gsk-your-actual-key-here"
$env:LLM_MODEL="llama-3.2-3b-preview"
```

Or add to `.env`:
```env
LLM_BASE_URL=https://api.groq.com/openai/v1
LLM_API_KEY=gsk_your_actual_key_here
LLM_MODEL=llama-3.2-3b-preview
```

### Step 5: Test

```powershell
curl https://api.groq.com/openai/v1/chat/completions `
  -H "Authorization: Bearer gsk_your_actual_key_here" `
  -H "Content-Type: application/json" `
  -d "{`"model`":`"llama-3.2-3b-preview`",`"messages`":[{`"role`":`"user`",`"content`":`"hi`"}]}"
```

---

## Hugging Face (Embeddings)

Hugging Face provides free inference API with rate limits:
- **~100 requests/hour** on free tier
- **Various embedding models** available
- **No credit card required**

### Step 1: Create Account

1. Go to [https://huggingface.co](https://huggingface.co)
2. Sign up with email, GitHub, or Google
3. Verify your email

### Step 2: Get API Key

1. Go to [https://huggingface.co/settings/tokens](https://huggingface.co/settings/tokens)
2. Click **"New token"**
3. Name: `ecommerce-ai-assistant`
4. Type: **Read** (sufficient for inference)
5. Click **"Generate token"**
6. Copy the token (starts with `hf_...`)

### Step 3: Choose Model

We recommend these embedding models:

| Model | Dimensions | Speed | Quality | Use Case |
|-------|-----------|-------|---------|----------|
| `sentence-transformers/all-MiniLM-L6-v2` | 384 | Fast | Good | General purpose, **recommended** |
| `sentence-transformers/all-mpnet-base-v2` | 768 | Medium | Better | Higher quality |
| `BAAI/bge-base-en-v1.5` | 768 | Medium | Better | English text |

**Recommendation:** `sentence-transformers/all-MiniLM-L6-v2` - fast, lightweight, good quality.

### Step 4: Configure

```powershell
# Set environment variables
$env:EMBEDDING_PROVIDER="huggingface"
$env:HUGGINGFACE_BASE_URL="https://api-inference.huggingface.co"
$env:HUGGINGFACE_EMBEDDING_MODEL="sentence-transformers/all-MiniLM-L6-v2"
$env:HUGGINGFACE_API_KEY="hf_your_actual_token_here"
```

Or add to `.env`:
```env
EMBEDDING_PROVIDER=huggingface
HUGGINGFACE_BASE_URL=https://api-inference.huggingface.co
HUGGINGFACE_EMBEDDING_MODEL=sentence-transformers/all-MiniLM-L6-v2
HUGGINGFACE_API_KEY=hf_your_actual_token_here
```

### Step 5: Test

```powershell
curl https://api-inference.huggingface.co/models/sentence-transformers/all-MiniLM-L6-v2 `
  -H "Authorization: Bearer hf_your_actual_token_here" `
  -H "Content-Type: application/json" `
  -d "{`"inputs`":`"test`"}"
```

---

## Configuration

### Full `.env` Example (Groq + Hugging Face)

```env
APP_ENV=local
HTTP_ADDR=:8080

# Groq LLM
LLM_BASE_URL=https://api.groq.com/openai/v1
LLM_API_KEY=gsk_your_groq_key
LLM_MODEL=llama-3.2-3b-preview

# Hugging Face Embeddings
EMBEDDING_PROVIDER=huggingface
HUGGINGFACE_BASE_URL=https://api-inference.huggingface.co
HUGGINGFACE_EMBEDDING_MODEL=sentence-transformers/all-MiniLM-L6-v2
HUGGINGFACE_API_KEY=hf_your_huggingface_token

# Optional: WhatsApp (leave empty for dev mode)
WHATSAPP_VERIFY_TOKEN=
WHATSAPP_APP_SECRET=
WHATSAPP_ACCESS_TOKEN=
WHATSAPP_PHONE_NUMBER_ID=
WHATSAPP_GRAPH_API_VERSION=v18.0
WHATSAPP_WEBHOOK_SIGNATURE_REQUIRED=false

# Tool and Retrieval
MAX_TOOL_CALLS=5
TOP_K=5
CHUNK_SIZE=1000
CHUNK_OVERLAP=200
REQUEST_TIMEOUT=30
LLM_TIMEOUT=60
```

### Running the App

```powershell
# 1. Install dependencies
go mod download

# 2. Run the server
make run
# or
go run ./cmd/server
```

The app will:
- Use Groq for LLM completions
- Use Hugging Face for embeddings
- Start in dev mode without WhatsApp credentials
- Serve on `http://localhost:8080`

### Testing the Setup

```powershell
# Health check
curl http://localhost:8080/healthz

# Dev chat (test the assistant)
curl -X POST http://localhost:8080/api/v1/chat `
  -H "Content-Type: application/json" `
  -d "{`"message`":`"What payment methods do you accept?`",`"customer_id`":`"CUST-001`"}"
```

---

## Rate Limits and Usage Tips

### Groq Free Tier
- 30 requests/minute
- 7,000 requests/day
- Resets daily at midnight UTC
- Monitor usage at: https://console.groq.com/usage

**Tips:**
- Cache responses when possible
- Use shorter prompts for development
- The tool-call loop may make multiple requests per user message

### Hugging Face Free Tier
- ~100 requests/hour
- No daily limit, but rate-limited per minute
- Monitor at: https://huggingface.co/settings/rate-limits

**Tips:**
- Embed documents once, cache the vectors
- The app caches embeddings in-memory during runtime
- Restarting the app re-embeds documents (by design)

---

## Troubleshooting

### "API key invalid"
- Double-check the key (no extra spaces)
- Ensure you're using the correct environment variable
- For Hugging Face: token must have "read" permission

### "Rate limit exceeded"
- Groq: Wait 1 minute (30 req/min limit)
- Hugging Face: Wait 1 hour (100 req/hour limit)
- Consider upgrading to paid tier if needed

### "Model not found"
- Verify model name spelling
- For Groq: use exact model names from console
- For Hugging Face: ensure model exists and is accessible

### "Connection timeout"
- Check internet connection
- Verify firewall allows outbound HTTPS
- Increase `LLM_TIMEOUT` or `REQUEST_TIMEOUT` in `.env`

---

## Next Steps

1. **Test with dev chat** - Use the `/api/v1/chat` endpoint to verify everything works
2. **Add test cases** - Write evaluation cases in `cmd/evals/main.go`
3. **Monitor usage** - Check Groq and Hugging Face dashboards for rate limits
4. **Upgrade if needed** - Both platforms have affordable paid tiers

---

## Alternative: Run Locally with Ollama

If cloud APIs are too slow or unreliable, run locally:

```powershell
# Install Ollama from https://ollama.ai
ollama pull llama3.2:3b
ollama pull nomic-embed-text

# Update .env
LLM_BASE_URL=http://localhost:11434/v1
LLM_API_KEY=ollama
LLM_MODEL=llama3.2:3b
EMBEDDING_PROVIDER=ollama
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_EMBEDDING_MODEL=nomic-embed-text
```

This requires ~8GB RAM but has no rate limits.
