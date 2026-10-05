package application

import (
	"context"
	"testing"

	catalogmemory "ecommerce-ai-assistant/internal/catalog/adapters/memory"
	knowledgememory "ecommerce-ai-assistant/internal/knowledge/adapters/memory"
	ordersmemory "ecommerce-ai-assistant/internal/orders/adapters/memory"
)

func TestToolRegistrySearchProducts(t *testing.T) {
	catalog := catalogmemory.NewProductRepository()
	orders := ordersmemory.NewOrderRepository()
	knowledge := knowledgememory.NewVectorRepository()
	registry := NewToolRegistry(catalog, orders, knowledge)

	args := map[string]any{"query": "hoodie"}
	result, err := registry.Execute(context.Background(), ToolCallRequest{
		CallID: "1", Tool: "search_products", Args: args,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
	if len(result.Content) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestToolRegistryUnknownTool(t *testing.T) {
	registry := NewToolRegistry(nil, nil, nil)
	_, err := registry.Execute(context.Background(), ToolCallRequest{
		CallID: "1", Tool: "unknown_tool", Args: map[string]any{},
	})
	if err == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestToolRegistryMissingArgs(t *testing.T) {
	catalog := catalogmemory.NewProductRepository()
	orders := ordersmemory.NewOrderRepository()
	knowledge := knowledgememory.NewVectorRepository()
	registry := NewToolRegistry(catalog, orders, knowledge)

	_, err := registry.Execute(context.Background(), ToolCallRequest{
		CallID: "1", Tool: "get_product", Args: map[string]any{},
	})
	if err == nil {
		t.Error("expected error for missing args")
	}
}
