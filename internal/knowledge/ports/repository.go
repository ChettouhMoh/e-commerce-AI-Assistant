// Package ports defines the repository interfaces for the knowledge domain.
package ports

import (
	"context"

	knowledgedomain "ecommerce-ai-assistant/internal/knowledge/domain"
)

// KnowledgeRepository defines the interface for knowledge base persistence.
type KnowledgeRepository interface {
	// Search performs a semantic search and returns the top chunks.
	Search(ctx context.Context, query string, topK int) ([]knowledgedomain.RetrievalResult, error)

	// Store persists policy chunks for later retrieval.
	Store(ctx context.Context, chunks []knowledgedomain.PolicyChunk) error
}
