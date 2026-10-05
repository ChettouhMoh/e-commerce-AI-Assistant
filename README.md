# E-commerce AI Assistant

A modular Go backend that answers customer questions through WhatsApp using LLM tool calling and RAG-based policy retrieval.

## Quick Start

```bash
cp .env.example .env
go mod download
make run
```

The app starts in local/mock mode. Use the dev chat endpoint to test without WhatsApp.

## Architecture

See `docs/architecture.md` for bounded context boundaries and dependency direction.

## Configuration

Copy `.env.example` to `.env` and set values. The app starts without external credentials in local mode.

## Testing

```bash
make test
make eval
```

## Documentation

- `docs/architecture.md` - System design
- `docs/data-flow.md` - Request lifecycle
- `docs/evaluation.md` - Evaluation guide
- `docs/setup-whatsapp.md` - WhatsApp setup
