// Package ports defines the interfaces for embedding providers.
package ports

import (
	"context"
)

// EmbeddingProvider defines the interface for generating text embeddings.
type EmbeddingProvider interface {
	// Embed generates an embedding vector for the given text.
	Embed(ctx context.Context, text string) ([]float64, error)

	// EmbedBatch generates embedding vectors for multiple texts.
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)

	// Dimension returns the dimensionality of the embedding vectors.
	Dimension() int
}
