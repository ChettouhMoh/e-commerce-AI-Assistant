package ports

import (
	"context"
)

// EmbeddingProvider abstracts the text-to-vector step so the RAG pipeline can
// swap between Ollama, OpenAI, or a fake implementation.
type EmbeddingProvider interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
	Dimension() int
}
