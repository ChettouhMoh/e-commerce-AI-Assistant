package memory_test

import (
	"context"
	"testing"

	"ecommerce-ai-assistant/internal/orders/adapters/memory"
	"ecommerce-ai-assistant/internal/orders/domain"
)

func TestOrderRepositoryGetByReference(t *testing.T) {
	repo := memory.NewOrderRepository()
	ctx := context.Background()

	t.Run("valid order and customer", func(t *testing.T) {
		o, err := repo.GetByReference(ctx, "ORD-1001", "CUST-001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if o.Status != domain.OrderStatusShipped {
			t.Errorf("expected shipped, got %s", o.Status)
		}
	})

	t.Run("wrong customer denied", func(t *testing.T) {
		_, err := repo.GetByReference(ctx, "ORD-1001", "CUST-002")
		if err == nil {
			t.Error("expected error for wrong customer")
		}
	})

	t.Run("missing order", func(t *testing.T) {
		_, err := repo.GetByReference(ctx, "NON-EXISTENT", "CUST-001")
		if err == nil {
			t.Error("expected error for missing order")
		}
	})
}

func TestOrderRepositoryListByCustomer(t *testing.T) {
	repo := memory.NewOrderRepository()
	ctx := context.Background()

	orders, err := repo.ListByCustomer(ctx, "CUST-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 3 {
		t.Errorf("expected 3 orders, got %d", len(orders))
	}
}
