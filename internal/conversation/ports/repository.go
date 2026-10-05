// Package ports defines the repository interface for the conversation domain.
package ports

import (
	"context"
	"time"

	convdomain "ecommerce-ai-assistant/internal/conversation/domain"
)

// ConversationRepository defines the interface for conversation persistence.
type ConversationRepository interface {
	// FindByID retrieves a conversation by its ID.
	FindByID(ctx context.Context, id string) (*convdomain.Conversation, error)

	// Save persists a conversation.
	Save(ctx context.Context, conv *convdomain.Conversation) error

	// FindByUserID retrieves conversations for a specific user.
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]convdomain.Conversation, error)

	// FindActiveByUserID retrieves the most recent active conversation for a user.
	FindActiveByUserID(ctx context.Context, userID, channelID string) (*convdomain.Conversation, error)

	// AppendTurn adds a new turn to an existing conversation.
	AppendTurn(ctx context.Context, conversationID string, turn *convdomain.Turn) error

	// Delete removes a conversation by ID.
	Delete(ctx context.Context, id string) error

	// DeleteOlderThan removes conversations older than the given time.
	DeleteOlderThan(ctx context.Context, olderThan time.Time) error
}
