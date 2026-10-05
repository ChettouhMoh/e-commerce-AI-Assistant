// Package conversation contains the core domain models for multi-turn conversations.
package conversation

import (
	"time"
)

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

// Turn represents a single turn in a conversation.
type Turn struct {
	ID                string
	ConversationID    string
	UserMessage       Message
	AssistantResponse Message
	ToolCalls         []ToolCall
	TokenUsage        TokenUsage
	CreatedAt         time.Time
}

// Message represents a single message in a conversation turn.
type Message struct {
	Role    string
	Content string
}

// ToolCall represents a tool call within a turn.
type ToolCall struct {
	ID        string
	Name      string
	Arguments map[string]interface{}
}

// TokenUsage represents token consumption for a turn.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}
