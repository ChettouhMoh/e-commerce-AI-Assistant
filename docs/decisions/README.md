# Decision Records

## 1. Gin and Modular Monolith

We use Gin only in transport/bootstrap packages. The modular monolith provides clear boundaries without operational complexity.

## 2. In-Memory Repositories

In-memory repositories are the initial persistence adapter. They require no external dependencies and make the project easy to run and test.

## 3. Embedding/Vector-Index Choice

We use an in-memory vector index with cosine similarity. An Ollama embedding adapter is provided for local use. If embeddings are unavailable, the app still starts in mock mode.

## 4. Provider and Channel Ports

Provider and channel interfaces allow fake adapters for local testing. Real adapters implement the same ports.

## 5. Why Structured Tools and RAG

Structured business facts (products, orders) use tools for precise queries. Policies use RAG for natural language retrieval. This separation keeps each system simple and testable.
