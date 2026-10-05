package memory_test

import (
	"context"
	"testing"

	"ecommerce-ai-assistant/internal/knowledge/adapters/memory"
	"ecommerce-ai-assistant/internal/knowledge/domain"
)

func TestVectorRepositorySearch(t *testing.T) {
	repo := memory.NewVectorRepository()
	ctx := context.Background()

	chunks := []domain.PolicyChunk{
		{ID: "1", Source: "return", Text: "Returns accepted within 30 days", Vector: []float64{1, 0, 0}},
		{ID: "2", Source: "return", Text: "Items must be unworn", Vector: []float64{0.9, 0.1, 0}},
		{ID: "3", Source: "shipping", Text: "Free shipping on orders over $50", Vector: []float64{0, 1, 0}},
	}
	if err := repo.Store(ctx, chunks); err != nil {
		t.Fatalf("store failed: %v", err)
	}

	results, err := repo.Search(ctx, "return policy", 2)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	if results[0].Score <= results[1].Score {
		t.Error("expected descending score order")
	}
}

func TestVectorRepositoryZeroVector(t *testing.T) {
	repo := memory.NewEmptyVectorRepository()
	ctx := context.Background()

	chunks := []domain.PolicyChunk{
		{ID: "1", Source: "test", Text: "test", Vector: []float64{}},
	}
	if err := repo.Store(ctx, chunks); err != nil {
		t.Fatalf("store failed: %v", err)
	}

	results, err := repo.Search(ctx, "test", 1)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty vector, got %d", len(results))
	}
}
