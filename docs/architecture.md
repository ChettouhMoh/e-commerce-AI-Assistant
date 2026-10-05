# Architecture

## Bounded Contexts

The project is organized into bounded contexts, each owning its domain concepts, use cases, ports, and adapters:

- `catalog` - Products, variants, inventory, and product search
- `orders` - Order status, customer association, and order lookup
- `knowledge` - Policy loading, chunking, embeddings, vector retrieval, and optional reranking
- `assistant` - Prompt construction, LLM orchestration, tool-call loop, and final response
- `tools` - Tool definitions, schemas, allowlisting, validation, and dispatch
- `conversation` - Bounded conversation history and inbound-message idempotency
- `channels` - Provider-neutral messaging contracts and concrete channel adapters

## Dependency Direction

Domain and application logic must not depend on Gin, Meta webhook payloads, LLM SDKs, or a specific storage implementation.

```
domain -> ports <- adapters
```

## Why a Modular Monolith

A modular monolith is appropriate for this learning project because it provides clear boundaries without the operational complexity of microservices. Each context can be extracted later if needed.
