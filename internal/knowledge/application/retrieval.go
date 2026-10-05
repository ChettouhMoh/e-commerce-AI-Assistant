package application

import (
	"context"
	"math"

	"ecommerce-ai-assistant/internal/knowledge/domain"
	"ecommerce-ai-assistant/internal/knowledge/ports"
)

type EmbeddingProvider = ports.EmbeddingProvider

type VectorStore interface {
	Upsert(ctx context.Context, chunks []domain.PolicyChunk) error
	Query(ctx context.Context, vector []float64, topK int) ([]domain.RetrievalResult, error)
}

type Retriever struct {
	embedder EmbeddingProvider
	store    VectorStore
	topK     int
}

func NewRetriever(embedder EmbeddingProvider, store VectorStore, topK int) *Retriever {
	if topK <= 0 {
		topK = 5
	}
	return &Retriever{embedder: embedder, store: store, topK: topK}
}

func (r *Retriever) Retrieve(ctx context.Context, query string) ([]domain.RetrievalResult, error) {
	embeddings, err := r.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 || isZeroVector(embeddings[0]) {
		return nil, nil
	}

	return r.store.Query(ctx, embeddings[0], r.topK)
}

func (r *Retriever) Ingest(ctx context.Context, chunks []Chunk) (int, error) {
	if len(chunks) == 0 {
		return 0, nil
	}

	texts := make([]string, len(chunks))
	for i, ch := range chunks {
		texts[i] = ch.Text
	}

	embeddings, err := r.embedder.Embed(ctx, texts)
	if err != nil {
		return 0, err
	}

	if len(embeddings) != len(chunks) {
		return 0, nil
	}

	var policyChunks []domain.PolicyChunk
	for i, ch := range chunks {
		vec := embeddings[i]
		if isZeroVector(vec) {
			continue
		}
		policyChunks = append(policyChunks, domain.PolicyChunk{
			ID:       ch.Source + "_" + string(rune('0'+i)),
			Source:   ch.Source,
			Text:     ch.Text,
			Vector:   vec,
			Metadata: map[string]string{"heading": ch.Heading},
		})
	}

	if len(policyChunks) == 0 {
		return 0, nil
	}

	if err := r.store.Upsert(ctx, policyChunks); err != nil {
		return 0, err
	}

	return len(policyChunks), nil
}

func isZeroVector(v []float64) bool {
	for _, f := range v {
		if f != 0 {
			return false
		}
	}
	return true
}

func CosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
