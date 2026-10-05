// Package assistant contains the core domain models for the AI assistant.
package domain

import "time"

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

// ToolCallRequest represents a request to execute a tool.
type ToolCallRequest struct {
	CallID string
	Tool   string
	Args   map[string]any
}

// ToolResultData represents the data returned by a tool execution.
type ToolResultData struct {
	Content string
	IsError bool
}

// ToolDefinition describes a tool that the LLM can call.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]ParameterDef
	Required    []string
}

// ParameterDef describes a single parameter of a tool.
type ParameterDef struct {
	Type        string
	Description string
	Enum        []string
}

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

// JSONSchema returns the JSON Schema representation of the tool definition.
func (t ToolDefinition) JSONSchema() map[string]any {
	props := make(map[string]any)
	for k, v := range t.Parameters {
		prop := map[string]any{
			"type":        v.Type,
			"description": v.Description,
		}
		if len(v.Enum) > 0 {
			prop["enum"] = v.Enum
		}
		props[k] = prop
	}
	return map[string]any{
		"type":       "object",
		"properties": props,
		"required":   t.Required,
	}
}
