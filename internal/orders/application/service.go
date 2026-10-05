package application

import (
	"context"

	ordersdomain "ecommerce-ai-assistant/internal/orders/domain"
	"ecommerce-ai-assistant/internal/orders/ports"
)

// OrdersService manages order operations.
type OrdersService struct {
	repo ports.OrderRepository
}

// NewOrdersService creates a new OrdersService.
func NewOrdersService(repo ports.OrderRepository) *OrdersService {
	return &OrdersService{repo: repo}
}

// GetOrderStatus retrieves the status of a specific order.
func (s *OrdersService) GetOrderStatus(ctx context.Context, reference, customerID string) (*ordersdomain.Order, error) {
	return s.repo.GetByReference(ctx, reference, customerID)
}

// GetCustomerOrders lists all orders for a given customer.
func (s *OrdersService) GetCustomerOrders(ctx context.Context, customerID string) ([]ordersdomain.Order, error) {
	return s.repo.ListByCustomer(ctx, customerID)
}
