package ports

import (
	"context"
	"ecommerce-ai-assistant/internal/orders/domain"
)

type OrderRepository interface {
	GetByReference(ctx context.Context, reference, customerID string) (*domain.Order, error)
	ListByCustomer(ctx context.Context, customerID string) ([]domain.Order, error)
}
