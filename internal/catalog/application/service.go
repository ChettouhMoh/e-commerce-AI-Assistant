package application

import (
	"context"
	"ecommerce-ai-assistant/internal/catalog/domain"
	"ecommerce-ai-assistant/internal/catalog/ports"
)

// CatalogService manages product catalog operations.
type CatalogService struct {
	repo ports.ProductRepository
}

// NewCatalogService creates a new CatalogService.
func NewCatalogService(repo ports.ProductRepository) *CatalogService {
	return &CatalogService{repo: repo}
}

// SearchProducts searches the product catalog.
func (s *CatalogService) SearchProducts(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	return s.repo.Search(ctx, filter)
}

// GetProduct retrieves a product by ID.
func (s *CatalogService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

// CheckInventory checks stock for a variant by SKU.
func (s *CatalogService) CheckInventory(ctx context.Context, sku string) (*domain.Variant, error) {
	return s.repo.GetVariant(ctx, sku)
}
