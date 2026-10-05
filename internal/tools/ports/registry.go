// Package ports defines the interfaces for the tools system.
package ports

import (
	"context"

	toolsdomain "ecommerce-ai-assistant/internal/tools/domain"
)

// ToolRegistry defines the interface for managing and executing tools.
type ToolRegistry interface {
	// List returns all available tool definitions.
	List() []toolsdomain.ToolDefinition

	// Execute invokes a tool by name with the given arguments.
	Execute(ctx context.Context, call toolsdomain.ToolCallRequest) (toolsdomain.ToolResultData, error)

	// ValidateArgs validates the arguments for a tool definition.
	ValidateArgs(def toolsdomain.ToolDefinition, args map[string]any) error
}
