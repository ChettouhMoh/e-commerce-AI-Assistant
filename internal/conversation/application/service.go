package application

import (
	"context"
	"time"

	convdomain "ecommerce-ai-assistant/internal/conversation/domain"
	convports "ecommerce-ai-assistant/internal/conversation/ports"
)

// ConversationService manages conversation lifecycle operations.
type ConversationService struct {
	repo convports.ConversationRepository
}

// NewConversationService creates a new ConversationService.
func NewConversationService(repo convports.ConversationRepository) *ConversationService {
	return &ConversationService{repo: repo}
}

// GetOrCreate retrieves an existing conversation or creates a new one if not found.
func (s *ConversationService) GetOrCreate(ctx context.Context, id string) (*convdomain.Conversation, error) {
	conv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		conv = &convdomain.Conversation{
			ID:        id,
			Turns:     []convdomain.Turn{},
			Metadata:  make(map[string]string),
			CreatedAt: time.Now(),
		}
	}
	return conv, nil
}

// Save persists a conversation.
func (s *ConversationService) Save(ctx context.Context, conv *convdomain.Conversation) error {
	conv.UpdatedAt = time.Now()
	return s.repo.Save(ctx, conv)
}
