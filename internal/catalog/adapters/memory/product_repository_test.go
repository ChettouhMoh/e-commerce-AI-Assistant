package memory_test

import (
	"context"
	"strings"
	"testing"

	"ecommerce-ai-assistant/internal/catalog/adapters/memory"
	"ecommerce-ai-assistant/internal/catalog/domain"
)

func TestProductRepositorySearch(t *testing.T) {
	repo := memory.NewProductRepository()
	ctx := context.Background()

	t.Run("search by query", func(t *testing.T) {
		results, err := repo.Search(ctx, domain.ProductFilter{Query: "hoodie", Limit: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) == 0 {
			t.Error("expected at least one result")
		}
	})

	t.Run("search by category", func(t *testing.T) {
		results, err := repo.Search(ctx, domain.ProductFilter{Category: "shoes", Limit: 10})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, p := range results {
			if p.Category != "shoes" {
				t.Errorf("expected category shoes, got %s", p.Category)
			}
		}
	})

	t.Run("filter by color and size", func(t *testing.T) {
		results, err := repo.Search(ctx, domain.ProductFilter{Color: "blue", Size: "M", Limit: 10})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, p := range results {
			for _, v := range p.Variants {
				if !strings.EqualFold(v.Color, "blue") {
					t.Errorf("expected color blue, got %s", v.Color)
				}
				if !strings.EqualFold(v.Size, "M") {
					t.Errorf("expected size M, got %s", v.Size)
				}
			}
		}
	})
}

func TestProductRepositoryGetByID(t *testing.T) {
	repo := memory.NewProductRepository()
	ctx := context.Background()

	p, err := repo.GetByID(ctx, "PROD-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != "PROD-001" {
		t.Errorf("expected PROD-001, got %s", p.ID)
	}

	_, err = repo.GetByID(ctx, "NON-EXISTENT")
	if err == nil {
		t.Error("expected error for missing product")
	}
}

func TestProductRepositoryGetVariant(t *testing.T) {
	repo := memory.NewProductRepository()
	ctx := context.Background()

	v, err := repo.GetVariant(ctx, "HOOD-BLU-M")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.SKU != "HOOD-BLU-M" {
		t.Errorf("expected HOOD-BLU-M, got %s", v.SKU)
	}
	if v.Stock != 8 {
		t.Errorf("expected stock 8, got %d", v.Stock)
	}
}
