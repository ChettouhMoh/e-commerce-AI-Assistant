// Package domain contains the core domain models for product catalog operations.
package domain

import "time"

// Product represents a product in the catalog.
type Product struct {
	ID          string
	Name        string
	Description string
	Category    string
	Price       float64
	Currency    string
	Variants    []Variant
	Images      []string
	Metadata    map[string]string
	CreatedAt   time.Time
}

// Variant represents a specific variant of a product (e.g., size, color).
type Variant struct {
	SKU       string
	ProductID string
	Color     string
	Size      string
	Stock     int
	Barcode   string
}

// Money represents a monetary amount with currency.
type Money struct {
	Amount   float64
	Currency string
}

// Stock represents inventory stock for a product variant.
type Stock struct {
	VariantID string
	Quantity  int
	Available int
	Location  string
}

// ProductFilter represents criteria for filtering products.
type ProductFilter struct {
	Query    string
	Category string
	Color    string
	Size     string
	MinPrice float64
	MaxPrice float64
	Limit    int
}

// Category constants for the flat string Category field.
const (
	CategoryHoodies     string = "hoodies"
	CategoryShoes       string = "shoes"
	CategoryPants       string = "pants"
	CategoryAccessories string = "accessories"
)
