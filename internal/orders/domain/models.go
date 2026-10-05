// Package domain contains the core domain models for order operations.
package domain

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusReturned  OrderStatus = "returned"
)

// IsValid returns true if the order status is a known valid status.
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending, OrderStatusPaid, OrderStatusShipped,
		OrderStatusDelivered, OrderStatusCancelled, OrderStatusReturned:
		return true
	}
	return false
}

// IsTerminal returns true if the order status indicates no further changes.
func (s OrderStatus) IsTerminal() bool {
	switch s {
	case OrderStatusDelivered, OrderStatusCancelled, OrderStatusReturned:
		return true
	}
	return false
}

// OrderItem represents a single item within an order.
type OrderItem struct {
	SKU      string
	Product  string
	Variant  string
	Quantity int
	Price    float64
}

// Money represents a monetary amount with currency.
type Money struct {
	Amount   float64
	Currency string
}

// Order represents a customer order.
type Order struct {
	ID             string
	CustomerID     string
	OrderNumber    string
	Status         OrderStatus
	Items          []OrderItem
	Total          float64
	Currency       string
	TrackingNumber string
	CreatedAt      string
	UpdatedAt      string
}

// ShippingInfo represents shipping details for an order.
type ShippingInfo struct {
	Address string
	City    string
	Country string
	Zip     string
	Method  string
}
