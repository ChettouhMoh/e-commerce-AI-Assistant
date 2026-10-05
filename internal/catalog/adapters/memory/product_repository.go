package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"ecommerce-ai-assistant/internal/catalog/domain"
	"ecommerce-ai-assistant/internal/catalog/ports"
)

type ProductRepository struct {
	products map[string]domain.Product
	variants map[string]domain.Variant
	mu       sync.RWMutex
}

var _ ports.ProductRepository = (*ProductRepository)(nil)

func NewProductRepository() *ProductRepository {
	r := &ProductRepository{
		products: make(map[string]domain.Product),
		variants: make(map[string]domain.Variant),
	}
	r.Seed()
	return r
}

func (r *ProductRepository) Search(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []domain.Product
	query := strings.ToLower(filter.Query)

	for _, p := range r.products {
		if filter.Category != "" && p.Category != filter.Category {
			continue
		}
		if filter.MaxPrice > 0 && p.Price > filter.MaxPrice {
			continue
		}
		if query != "" {
			matched := strings.Contains(strings.ToLower(p.Name), query) || strings.Contains(strings.ToLower(p.Description), query)
			if !matched {
				continue
			}
		}
		productCopy := p
		filteredVariants := make([]domain.Variant, 0, len(p.Variants))
		for _, v := range p.Variants {
			if filter.Color != "" && !strings.EqualFold(v.Color, filter.Color) {
				continue
			}
			if filter.Size != "" && !strings.EqualFold(v.Size, filter.Size) {
				continue
			}
			filteredVariants = append(filteredVariants, v)
		}
		productCopy.Variants = filteredVariants
		results = append(results, productCopy)
	}

	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}
	return results, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.products[id]
	if !ok {
		return nil, fmt.Errorf("product not found: %s", id)
	}
	return &p, nil
}

func (r *ProductRepository) GetVariant(ctx context.Context, sku string) (*domain.Variant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.variants[sku]
	if !ok {
		return nil, fmt.Errorf("variant not found: %s", sku)
	}
	return &v, nil
}

func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.products[product.ID] = *product
	for _, v := range product.Variants {
		r.variants[v.SKU] = v
	}
	return nil
}

func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	return r.Create(ctx, product)
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.products, id)
	return nil
}

func (r *ProductRepository) Seed() {
	products := []domain.Product{
		{ID: "PROD-001", Name: "Classic Blue Hoodie", Description: "Comfortable cotton hoodie in blue", Category: string(domain.CategoryHoodies), Price: 59.99, Currency: "USD", CreatedAt: time.Now().Add(-120 * time.Hour), Variants: []domain.Variant{
			{SKU: "HOOD-BLU-S", ProductID: "PROD-001", Color: "blue", Size: "S", Stock: 15, Barcode: "BC-HOOD-BLU-S"},
			{SKU: "HOOD-BLU-M", ProductID: "PROD-001", Color: "blue", Size: "M", Stock: 8, Barcode: "BC-HOOD-BLU-M"},
			{SKU: "HOOD-BLU-L", ProductID: "PROD-001", Color: "blue", Size: "L", Stock: 3, Barcode: "BC-HOOD-BLU-L"},
		}},
		{ID: "PROD-002", Name: "Running Shoes Pro", Description: "Lightweight running shoes", Category: string(domain.CategoryShoes), Price: 129.99, Currency: "USD", CreatedAt: time.Now().Add(-110 * time.Hour), Variants: []domain.Variant{
			{SKU: "SHOE-RUN-BLK-9", ProductID: "PROD-002", Color: "black", Size: "9", Stock: 20, Barcode: "BC-SHOE-RUN-BLK-9"},
			{SKU: "SHOE-RUN-BLK-10", ProductID: "PROD-002", Color: "black", Size: "10", Stock: 12, Barcode: "BC-SHOE-RUN-BLK-10"},
			{SKU: "SHOE-RUN-WHT-9", ProductID: "PROD-002", Color: "white", Size: "9", Stock: 5, Barcode: "BC-SHOE-RUN-WHT-9"},
		}},
		{ID: "PROD-003", Name: "Casual Chinos", Description: "Slim fit casual chinos", Category: string(domain.CategoryPants), Price: 49.99, Currency: "USD", CreatedAt: time.Now().Add(-100 * time.Hour), Variants: []domain.Variant{
			{SKU: "CHINO-30", ProductID: "PROD-003", Color: "beige", Size: "30", Stock: 25, Barcode: "BC-CHINO-30"},
			{SKU: "CHINO-32", ProductID: "PROD-003", Color: "beige", Size: "32", Stock: 18, Barcode: "BC-CHINO-32"},
		}},
		{ID: "PROD-004", Name: "Wool Scarf", Description: "Soft wool scarf", Category: string(domain.CategoryAccessories), Price: 29.99, Currency: "USD", CreatedAt: time.Now().Add(-90 * time.Hour), Variants: []domain.Variant{
			{SKU: "SCARF-GRY", ProductID: "PROD-004", Color: "gray", Size: "One Size", Stock: 50, Barcode: "BC-SCARF-GRY"},
			{SKU: "SCARF-RED", ProductID: "PROD-004", Color: "red", Size: "One Size", Stock: 30, Barcode: "BC-SCARF-RED"},
		}},
		{ID: "PROD-005", Name: "Denim Jacket", Description: "Classic denim jacket", Category: string(domain.CategoryHoodies), Price: 89.99, Currency: "USD", CreatedAt: time.Now().Add(-80 * time.Hour), Variants: []domain.Variant{
			{SKU: "JACK-DENIM-M", ProductID: "PROD-005", Color: "blue", Size: "M", Stock: 7, Barcode: "BC-JACK-DENIM-M"},
			{SKU: "JACK-DENIM-L", ProductID: "PROD-005", Color: "blue", Size: "L", Stock: 4, Barcode: "BC-JACK-DENIM-L"},
		}},
		{ID: "PROD-006", Name: "Sneakers Urban", Description: "Urban style sneakers", Category: string(domain.CategoryShoes), Price: 79.99, Currency: "USD", CreatedAt: time.Now().Add(-70 * time.Hour), Variants: []domain.Variant{
			{SKU: "SNK-URB-BLK-8", ProductID: "PROD-006", Color: "black", Size: "8", Stock: 0, Barcode: "BC-SNK-URB-BLK-8"},
			{SKU: "SNK-URB-WHT-9", ProductID: "PROD-006", Color: "white", Size: "9", Stock: 10, Barcode: "BC-SNK-URB-WHT-9"},
		}},
		{ID: "PROD-007", Name: "Linen T-Shirt", Description: "Breathable linen t-shirt", Category: string(domain.CategoryPants), Price: 34.99, Currency: "USD", CreatedAt: time.Now().Add(-60 * time.Hour), Variants: []domain.Variant{
			{SKU: "TSHIRT-WHT-M", ProductID: "PROD-007", Color: "white", Size: "M", Stock: 40, Barcode: "BC-TSHIRT-WHT-M"},
			{SKU: "TSHIRT-BLK-L", ProductID: "PROD-007", Color: "black", Size: "L", Stock: 22, Barcode: "BC-TSHIRT-BLK-L"},
		}},
		{ID: "PROD-008", Name: "Winter Gloves", Description: "Warm winter gloves", Category: string(domain.CategoryAccessories), Price: 24.99, Currency: "USD", CreatedAt: time.Now().Add(-50 * time.Hour), Variants: []domain.Variant{
			{SKU: "GLOV-BLK", ProductID: "PROD-008", Color: "black", Size: "One Size", Stock: 60, Barcode: "BC-GLOV-BLK"},
		}},
		{ID: "PROD-009", Name: "Canvas Backpack", Description: "Durable canvas backpack for daily use", Category: string(domain.CategoryAccessories), Price: 64.99, Currency: "USD", CreatedAt: time.Now().Add(-40 * time.Hour), Variants: []domain.Variant{
			{SKU: "BAG-CAN-OLV", ProductID: "PROD-009", Color: "olive", Size: "One Size", Stock: 18, Barcode: "BC-BAG-CAN-OLV"},
			{SKU: "BAG-CAN-BLK", ProductID: "PROD-009", Color: "black", Size: "One Size", Stock: 25, Barcode: "BC-BAG-CAN-BLK"},
		}},
		{ID: "PROD-010", Name: "Slim Fit Jeans", Description: "Stretch denim slim fit jeans", Category: string(domain.CategoryPants), Price: 69.99, Currency: "USD", CreatedAt: time.Now().Add(-30 * time.Hour), Variants: []domain.Variant{
			{SKU: "JEAN-30", ProductID: "PROD-010", Color: "indigo", Size: "30", Stock: 14, Barcode: "BC-JEAN-30"},
			{SKU: "JEAN-32", ProductID: "PROD-010", Color: "indigo", Size: "32", Stock: 9, Barcode: "BC-JEAN-32"},
			{SKU: "JEAN-34", ProductID: "PROD-010", Color: "indigo", Size: "34", Stock: 6, Barcode: "BC-JEAN-34"},
		}},
	}

	for _, p := range products {
		r.products[p.ID] = p
		for _, v := range p.Variants {
			r.variants[v.SKU] = v
		}
	}
}
