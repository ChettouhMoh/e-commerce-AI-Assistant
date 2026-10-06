// Package assistant contains the core domain models for the AI assistant.
package domain

import (
	"time"

	toolsdomain "ecommerce-ai-assistant/internal/tools/domain"
)

// Role represents the role of a message sender.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// Message represents a message in the assistant conversation.
type Message struct {
	Role    Role
	Content string
}

// ToolCall represents a request to invoke a tool.
type ToolCall struct {
	ID        string
	Name      string
	Arguments map[string]any
}

// ToolResult represents the result of a tool invocation.
type ToolResult struct {
	CallID  string
	Content string
	IsError bool
}

// ToolDefinition is an alias for the tools domain ToolDefinition.
type ToolDefinition = toolsdomain.ToolDefinition

// ParameterDef is an alias for the tools domain ParameterDef.
type ParameterDef = toolsdomain.ParameterDef

// ToolCallRequest is an alias for the tools domain ToolCallRequest.
type ToolCallRequest = toolsdomain.ToolCallRequest

// ToolResultData is an alias for the tools domain ToolResultData.
type ToolResultData = toolsdomain.ToolResultData

// AssistantResponse represents the full response from the assistant.
type AssistantResponse struct {
	Messages   []Message
	ToolCalls  []ToolCall
	StopReason string
	Usage      Usage
	Model      string
}

// Usage represents token consumption metrics.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// Turn represents a single turn in a conversation.
type Turn struct {
	ID          string
	Role        Role
	Content     string
	ToolCalls   []ToolCall
	ToolResults []ToolResult
	CreatedAt   time.Time
}

// Conversation represents a multi-turn conversation session.
type Conversation struct {
	ID        string
	ChannelID string
	UserID    string
	Turns     []Turn
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}
