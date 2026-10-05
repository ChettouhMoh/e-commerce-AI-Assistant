# Data Flow

## Incoming WhatsApp Message

1. Meta sends webhook to `POST /webhooks/whatsapp`
2. Signature is validated using HMAC-SHA256
3. Message ID is deduplicated
4. Text is normalized into an internal `InboundMessage`
5. Assistant use case is invoked
6. Response is sent back via WhatsApp Graph API

## LLM Tool-Call Loop

1. User message is added to conversation
2. LLM is called with available tools
3. If LLM returns tool calls, they are validated and executed
4. Results are fed back as observation messages
5. Loop continues until LLM returns final text or max calls is reached

## RAG Retrieval

1. Policy documents are chunked with heading/paragraph awareness
2. Embeddings are generated via configured provider
3. Vectors are stored in in-memory index
4. At query time, question is embedded and cosine similarity is computed
5. Top-k results are returned to LLM

## Error Handling

- All errors are wrapped with context
- External timeouts are respected
- Missing evidence is reported, not fabricated
- Customer ownership is verified before revealing order details
