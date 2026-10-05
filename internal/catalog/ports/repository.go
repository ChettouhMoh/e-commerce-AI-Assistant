package ports

import (
	"context"
	"ecommerce-ai-assistant/internal/catalog/domain"
)

type ProductRepository interface {
	Search(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error)
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	GetVariant(ctx context.Context, sku string) (*domain.Variant, error)
}
