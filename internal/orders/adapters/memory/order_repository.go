package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"ecommerce-ai-assistant/internal/orders/domain"
	"ecommerce-ai-assistant/internal/orders/ports"
)

type OrderRepository struct {
	orders map[string]domain.Order
	byCust map[string][]string
	mu     sync.RWMutex
}

var _ ports.OrderRepository = (*OrderRepository)(nil)

func NewOrderRepository() *OrderRepository {
	r := &OrderRepository{
		orders: make(map[string]domain.Order),
		byCust: make(map[string][]string),
	}
	r.Seed()
	return r
}

func (r *OrderRepository) GetByReference(ctx context.Context, reference, customerID string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, o := range r.orders {
		if o.ID == reference && o.CustomerID == customerID {
			copy := o
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("order not found")
}

func (r *OrderRepository) ListByCustomer(ctx context.Context, customerID string) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids, ok := r.byCust[customerID]
	if !ok {
		return nil, nil
	}
	var orders []domain.Order
	for _, id := range ids {
		orders = append(orders, r.orders[id])
	}
	sort.Slice(orders, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, orders[i].CreatedAt)
		tj, _ := time.Parse(time.RFC3339, orders[j].CreatedAt)
		return ti.After(tj)
	})
	return orders, nil
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = *order
	r.byCust[order.CustomerID] = append(r.byCust[order.CustomerID], order.ID)
	sort.Strings(r.byCust[order.CustomerID])
	return nil
}

func (r *OrderRepository) Seed() {
	now := time.Now().Format(time.RFC3339)
	orders := []domain.Order{
		{ID: "ORD-1001", CustomerID: "CUST-001", OrderNumber: "ORD-1001", Status: domain.OrderStatusShipped, Total: 189.98, Currency: "USD",
			Items:          []domain.OrderItem{{SKU: "HOOD-BLU-M", Product: "Classic Blue Hoodie", Variant: "blue/M", Quantity: 1, Price: 59.99}},
			TrackingNumber: "TRK-1001-ABC", CreatedAt: time.Now().Add(-72 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339)},
		{ID: "ORD-1002", CustomerID: "CUST-001", OrderNumber: "ORD-1002", Status: domain.OrderStatusDelivered, Total: 129.99, Currency: "USD",
			Items:          []domain.OrderItem{{SKU: "SHOE-RUN-BLK-10", Product: "Running Shoes Pro", Variant: "black/10", Quantity: 1, Price: 129.99}},
			TrackingNumber: "TRK-1002-XYZ", CreatedAt: time.Now().Add(-14 * 24 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Add(-10 * 24 * time.Hour).Format(time.RFC3339)},
		{ID: "ORD-1003", CustomerID: "CUST-002", OrderNumber: "ORD-1003", Status: domain.OrderStatusPending, Total: 84.98, Currency: "USD",
			Items:     []domain.OrderItem{{SKU: "CHINO-32", Product: "Casual Chinos", Variant: "beige/32", Quantity: 1, Price: 49.99}, {SKU: "TSHIRT-WHT-M", Product: "Linen T-Shirt", Variant: "white/M", Quantity: 1, Price: 34.99}},
			CreatedAt: time.Now().Add(-2 * time.Hour).Format(time.RFC3339), UpdatedAt: now},
		{ID: "ORD-1004", CustomerID: "CUST-003", OrderNumber: "ORD-1004", Status: domain.OrderStatusPaid, Total: 59.98, Currency: "USD",
			Items:     []domain.OrderItem{{SKU: "SCARF-RED", Product: "Wool Scarf", Variant: "red", Quantity: 2, Price: 29.99}},
			CreatedAt: time.Now().Add(-6 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Add(-5 * time.Hour).Format(time.RFC3339)},
		{ID: "ORD-1005", CustomerID: "CUST-002", OrderNumber: "ORD-1005", Status: domain.OrderStatusCancelled, Total: 89.99, Currency: "USD",
			Items:     []domain.OrderItem{{SKU: "JACK-DENIM-M", Product: "Denim Jacket", Variant: "blue/M", Quantity: 1, Price: 89.99}},
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Add(-28 * 24 * time.Hour).Format(time.RFC3339)},
		{ID: "ORD-1006", CustomerID: "CUST-001", OrderNumber: "ORD-1006", Status: domain.OrderStatusReturned, Total: 64.98, Currency: "USD",
			Items:     []domain.OrderItem{{SKU: "GLOV-BLK", Product: "Winter Gloves", Variant: "black", Quantity: 2, Price: 24.99}, {SKU: "TSHIRT-BLK-L", Product: "Linen T-Shirt", Variant: "black/L", Quantity: 1, Price: 14.99}},
			CreatedAt: time.Now().Add(-21 * 24 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Add(-18 * 24 * time.Hour).Format(time.RFC3339)},
		{ID: "ORD-1007", CustomerID: "CUST-003", OrderNumber: "ORD-1007", Status: domain.OrderStatusShipped, Total: 144.98, Currency: "USD",
			Items:          []domain.OrderItem{{SKU: "SNK-URB-WHT-9", Product: "Sneakers Urban", Variant: "white/9", Quantity: 1, Price: 79.99}, {SKU: "BAG-CAN-BLK", Product: "Canvas Backpack", Variant: "black", Quantity: 1, Price: 64.99}},
			TrackingNumber: "TRK-1007-DEF", CreatedAt: time.Now().Add(-48 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Add(-36 * time.Hour).Format(time.RFC3339)},
	}

	for _, o := range orders {
		r.orders[o.ID] = o
		r.byCust[o.CustomerID] = append(r.byCust[o.CustomerID], o.ID)
	}

	for id := range r.byCust {
		sort.Strings(r.byCust[id])
	}
}
