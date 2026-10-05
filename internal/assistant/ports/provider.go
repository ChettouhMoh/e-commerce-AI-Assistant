// Package ports defines the interfaces for the AI assistant.
package ports

import (
	"context"

	"ecommerce-ai-assistant/internal/assistant/domain"
)

// LLMProvider defines the interface for interacting with a large language model.
type LLMProvider interface {
	// Complete sends messages and tool definitions to the LLM and returns the response.
	Complete(ctx context.Context, messages []domain.Message, tools []domain.ToolDefinition) (*domain.AssistantResponse, error)
}

// ToolRegistry defines the interface for managing and executing tools.
type ToolRegistry interface {
	// List returns all available tool definitions.
	List() []domain.ToolDefinition

	// Execute invokes a tool by name with the given arguments.
	Execute(ctx context.Context, call domain.ToolCallRequest) (domain.ToolResultData, error)

	// ValidateArgs validates the arguments for a tool definition.
	ValidateArgs(def domain.ToolDefinition, args map[string]any) error
}
